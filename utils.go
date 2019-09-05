package gourmet

import (
    "errors"
    "github.com/google/gopacket"
    "github.com/google/gopacket/pcap"
    "log"
    "strconv"
)

func getTcpProtocol(transport gopacket.Flow) Protocol {
    a, _ := strconv.ParseUint(transport.Src().String(), 10, 16)
    b, _ := strconv.ParseUint(transport.Dst().String(), 10, 16)
    if a < b {
        return tcpProtocols[uint16(a)]
    }
    return tcpProtocols[uint16(b)]
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