package gourmet

import (
	"errors"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"log"
	"time"
)

type interfaceType byte

const (
	// packet capture types
	AfpacketType interfaceType = 1
	PfringType   interfaceType = 2
	LibpcapType  interfaceType = 3
)

type SensorOptions struct {
	InterfaceName string
	InterfaceType interfaceType
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
	intType       interfaceType
	streamFactory *TcpStreamFactory
}

func Start(options *SensorOptions) {
	var err error
	logger, err = newLogger("gourmet.log")
	if err != nil {
		log.Fatal(err)
	}
	s := &Sensor{
		streamFactory: &TcpStreamFactory{},
	}
	if options.InterfaceType == PfringType {
		s.source, err = newPfringSensor(options)
		if err != nil {
			fmt.Println(err)
			return
		}
		s.intType = PfringType
	} else if options.InterfaceType == AfpacketType {
		s.source, err = newAfpacketSensor(options)
		if err != nil {
			fmt.Println(err)
			return
		}
		s.intType = AfpacketType
	} else if options.InterfaceType == LibpcapType {
		s.source, err = newLibpcapSensor(options)
		if err != nil {
			fmt.Println(err)
			return
		}
		s.intType = LibpcapType
	} else {
		fmt.Println("interface type is not set")
		return
	}
	s.run()
}

func (s *Sensor) run() {
	s.streamFactory.createAssembler()
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
	if packet.TransportLayer() != nil {
		switch packet.TransportLayer().LayerType() {
		case layers.LayerTypeTCP:
			go s.streamFactory.newPacket(packet.NetworkLayer().NetworkFlow(), packet.TransportLayer().(*layers.TCP))
		}
	}
	if packet.ApplicationLayer() != nil {
		switch packet.ApplicationLayer().LayerType() {
		case layers.LayerTypeTLS:
			go processTlsPacket(packet)
		case layers.LayerTypeDNS:
			go processDnsPacket(packet)
		default:
		}
	}
}
