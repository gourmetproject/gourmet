package gourmet

import (
    "errors"
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
