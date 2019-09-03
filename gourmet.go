package gourmet

import (
    "errors"
    "github.com/google/gopacket/pcap"
    "log"
)

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