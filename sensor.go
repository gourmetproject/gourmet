package gourmet

import (
    "errors"
    "github.com/google/gopacket"
    "github.com/google/gopacket/layers"
    "github.com/google/gopacket/reassembly"
    "time"
)

type InterfaceType byte

const (
    AfpacketType InterfaceType = 0
    PfringType   InterfaceType = 1
    LibpcapType  InterfaceType = 2
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
    packetSource *gopacket.PacketSource
    assembler *reassembly.Assembler
    streams chan *TcpStream
}

func NewSensor(options *SensorOptions) (s *Sensor, err error) {
    var src *gopacket.PacketSource
    if options.InterfaceType == PfringType {
        src, err = newPfringSensor(options)
        if err != nil {
            return nil, err
        }
    } else if options.InterfaceType == AfpacketType {
        src, err = newAfpacketSensor(options)
        if err != nil {
            return nil, err
        }
    } else if options.InterfaceType == LibpcapType {
        src, err = newLibpcapSensor(options)
        if err != nil {
            return nil, err
        }
    } else {
        return nil, errors.New("interface type is not set")
    }
    return &Sensor {
        packetSource: src,
    }, nil
}

func (s *Sensor) GetPackets() chan gopacket.Packet {
    return s.packetSource.Packets()
}

// private implementation of reassembly.AssemblerContext interface
type assemblerContext struct { CaptureInfo gopacket.CaptureInfo }
func (ac *assemblerContext) GetCaptureInfo() gopacket.CaptureInfo { return ac.CaptureInfo }

// Make sure to call CloseTcpStreams when you no longer need to consume reassembled tcp streams
func (s *Sensor) GetTcpStreams() chan *TcpStream {
    s.streams = make(chan *TcpStream)
    factory := &tcpStreamFactory {
        sensor: s,
    }
    s.assembler = createAssembler(factory)
    ticker := time.Tick(time.Minute)
    go func() {
        for {
            select {
            case packet := <- s.GetPackets():
                var tcp *layers.TCP
                if packet.NetworkLayer() == nil || packet.TransportLayer() == nil ||
                    packet.TransportLayer().LayerType() != layers.LayerTypeTCP {
                    continue
                }
                tcp = packet.TransportLayer().(*layers.TCP)
                context := &assemblerContext {
                    CaptureInfo: packet.Metadata().CaptureInfo,
                }
                s.assembler.AssembleWithContext(packet.NetworkLayer().NetworkFlow(), tcp, context)
            case <- ticker:
                s.assembler.FlushCloseOlderThan(time.Now().Add(time.Minute * -2))
            }
        }
    }()
    return s.streams
}

func (s *Sensor) CloseTcpStreams() {
    s.assembler.FlushAll()
    close(s.streams)
}