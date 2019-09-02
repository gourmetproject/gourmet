package gourmet

import (
    "errors"
    "github.com/google/gopacket"
    "github.com/google/gopacket/pcap"
    "log"
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
    packets chan gopacket.Packet
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
        packets: src.Packets(),
    }, nil
}

func (s *Sensor) GetPackets() chan gopacket.Packet {
    return s.packets
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

func checkIfInterfaceExists(iface string) error {
    devices, err := pcap.FindAllDevs()
    if err != nil {
        log.Fatal(err)
    }
    for _, device := range devices {
        if device.Name == iface {
            return nil
        }
    }
    return errors.New("specified network interface does not exist")
}