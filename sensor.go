package gourmet

import (
    "errors"
    "github.com/google/gopacket"
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
    sourceInUse bool
    streamFactory *tcpStreamFactory
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
    sf := &tcpStreamFactory {
        streams:     make(chan *TcpStream),
    }
    s = &Sensor {
        packetSource:  src,
        streamFactory: sf,
    }
    return s, nil
}

func (s *Sensor) NewPacketsChannel() (chan gopacket.Packet, error) {
    if s.sourceInUse {
        return nil, errors.New("sensor already in use")
    }
    s.sourceInUse = true
    return s.packetSource.Packets(), nil
}

func (s *Sensor) ClosePacketsChannel() {
    close(s.packetSource.Packets())
    s.sourceInUse = false
}

func (s *Sensor) NewTcpStreamsChannel() (chan *TcpStream, error) {
    if s.sourceInUse {
        return nil, errors.New("sensor already in use")
    }
    go s.streamFactory.getTcpStreams(s.packetSource.Packets())
    return s.streamFactory.streams, nil
}

func (s *Sensor) CloseTcpStreamsChannel() {
    close(s.packetSource.Packets())
    close(s.streamFactory.streams)
    s.sourceInUse = false
}