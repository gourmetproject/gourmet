package gourmet

import (
	"errors"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"log"
	"net"
	"strconv"
)

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

func getInterfaceAddresses(interfaceName string) (addresses []string) {
	i, err := net.InterfaceByName(interfaceName)
	if err != nil {
		// this should never happen. If it does, our sensor is broken and needs to die...
		panic("interface invalid")
	}
	addrs, err := i.Addrs()
	if err != nil {
		panic("interface invalid")
	}
	for _, addr := range addrs {
		addresses = append(addresses, addr.String())
	}
	return addresses
}

func processPorts(transport gopacket.Flow) (srcPort, dstPort int) {
	srcPort, _ = strconv.Atoi(transport.Src().String())
	dstPort, _ = strconv.Atoi(transport.Dst().String())
	return srcPort, dstPort
}
