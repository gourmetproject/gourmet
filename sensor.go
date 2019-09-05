package gourmet

import (
    "errors"
    "fmt"
    "github.com/google/gopacket"
    "github.com/google/gopacket/layers"
    "log"
    "time"
)

type InterfaceType byte

const (
    AfpacketType InterfaceType = 1
    PfringType   InterfaceType = 2
    LibpcapType  InterfaceType = 3
)

type SensorOptions struct {
    InterfaceName string
    InterfaceType
    IsPromiscuous bool
    SnapLen       uint32
    Filter        string
    Timeout       int
}

func initOptions(opt *SensorOptions) error {
    if opt.InterfaceName == "" {
        return errors.New("interface not set in options")
    }
    err := checkIfInterfaceExists(opt.InterfaceName)
    if err != nil {
        return err
    }
    if opt.SnapLen == 0 {
        opt.SnapLen = 65536
    }
    return nil
}

type Sensor struct {
    source        gopacket.PacketDataSource
    packets       chan gopacket.Packet
    intType       InterfaceType
    streamFactory *tcpStreamFactory
}

func NewSensor(options *SensorOptions) (s *Sensor, err error) {
    sf := &tcpStreamFactory {
        streams:     make(chan *TcpStream),
    }
    s = &Sensor {
        streamFactory: sf,
    }
    if options.InterfaceType == PfringType {
        s.source, err = newPfringSensor(options)
        if err != nil {
            return nil, err
        }
        s.intType = PfringType
    } else if options.InterfaceType == AfpacketType {
        s.source, err = newAfpacketSensor(options)
        if err != nil {
            return nil, err
        }
        s.intType = AfpacketType
    } else if options.InterfaceType == LibpcapType {
        s.source, err = newLibpcapSensor(options)
        if err != nil {
            return nil, err
        }
        s.intType = LibpcapType
    } else {
        return nil, errors.New("interface type is not set")
    }
    s.packets = make(chan gopacket.Packet)
    s.streamFactory.streams = make(chan *TcpStream)
    s.streamFactory.createAssembler()
    go s.processPackets()
    return s, nil
}

func (s *Sensor) processPackets() {
    s.streamFactory.ticker = time.NewTicker(time.Second * 10)
    for {
        p, _, err := s.source.ReadPacketData()
        if err != nil {
            log.Println(err)
            continue
        }
        packet := gopacket.NewPacket(p, layers.LayerTypeEthernet, gopacket.Default)
        s.processPacket(packet)
    }
}

func (s *Sensor) processPacket(packet gopacket.Packet) {
    if packet.TransportLayer() != nil && packet.TransportLayer().LayerType() == layers.LayerTypeTCP {
        s.streamFactory.assemblePacket(packet.NetworkLayer().NetworkFlow(), packet.TransportLayer().(*layers.TCP))
    }
}

func (s *Sensor) Packets() (packets chan gopacket.Packet) {
    return s.packets
}

func (s *Sensor) Streams() (streams chan *TcpStream) {
    return s.streamFactory.streams
}

func ExampleSensor() {
    opt := &SensorOptions{
        InterfaceName: "wlan0",
        InterfaceType: AfpacketType,
        IsPromiscuous: true,
    }
    src, err := NewSensor(opt)
    if err != nil {
        log.Fatal(err)
    }
    counter := 0
    for stream := range src.Streams() {
        fmt.Printf("new stream: %s %s %d\n", stream.NetworkFlow().String(), stream.TransportFlow().String(), len(stream.Payload()))
        counter++
        if counter == 10 {
            break
        }
    }
}
