package gourmet

import (
    "errors"
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
    InterfaceType string
    IsPromiscuous bool
    SnapLength    uint32
    Filter        string
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