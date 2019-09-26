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
	// packet capture types
	AfpacketType InterfaceType = 1
	PfringType   InterfaceType = 2
	LibpcapType  InterfaceType = 3
)

type SensorMetadata struct {
	// The network interface that the sensor is capturing traffic
	NetworkInterface string
	// The IP address of the capturing network interface
	NetworkAddress   []string
}

func getSensorMetadata(interfaceName string) *SensorMetadata{
	return &SensorMetadata{
		NetworkInterface: interfaceName,
		NetworkAddress:   getInterfaceAddresses(interfaceName),
	}
}

type SensorOptions struct {
	InterfaceName string
	InterfaceType InterfaceType
	IsPromiscuous bool
	SnapLen       uint32
	Bpf           string
	LogFileName   string
	Analyzers     []Analyzer
}

var analyzers []Analyzer

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

type sensor struct {
	source        gopacket.PacketDataSource
	streamFactory *tcpStreamFactory
	connections   chan *Connection
}

func Start(options *SensorOptions) {
	var err error
	logger, err = newLogger(options.LogFileName, options.InterfaceName)
	if err != nil {
		log.Fatal(err)
	}
	c := make(chan *Connection)
	s := &sensor{
		connections: c,
		streamFactory: &tcpStreamFactory{
			connections: c,
		},
	}
	analyzers = options.Analyzers
	err = s.getPacketSource(options)
	if err != nil {
		log.Fatal(err)
	}
	go s.processConnections()
	fmt.Printf("\nGourmet is running and logging to %s. Press CTL+C to stop...", logger.fileName)
	s.run()
}

func (s *sensor) getPacketSource(options *SensorOptions) (err error) {
	if options.InterfaceType == PfringType {
		s.source, err = newPfringSensor(options)
		if err != nil {
			return err
		}
	} else if options.InterfaceType == AfpacketType {
		s.source, err = newAfpacketSensor(options)
		if err != nil {
			return err
		}
	} else if options.InterfaceType == LibpcapType {
		s.source, err = newLibpcapSensor(options)
		if err != nil {
			return err
		}
	} else {
		return errors.New("interface type is not set")
	}
	return nil
}

func (s *sensor) run() {
	s.streamFactory.createAssembler()
	s.streamFactory.ticker = time.NewTicker(time.Second * 10)
	for {
		p, ci, err := s.source.ReadPacketData()
		if err != nil {
			log.Println(err)
			continue
		}
		packet := gopacket.NewPacket(p, layers.LayerTypeEthernet, gopacket.DecodeStreamsAsDatagrams)
		go s.processNewPacket(packet, ci)
	}
}

func (s *sensor) processNewPacket(packet gopacket.Packet, ci gopacket.CaptureInfo) {
    if packet.TransportLayer() != nil {
    	layer := packet.TransportLayer()
		switch layer.LayerType() {
		case layers.LayerTypeTCP:
			s.streamFactory.newPacket(packet.NetworkLayer().NetworkFlow(), packet.TransportLayer().(*layers.TCP))
			return
		case layers.LayerTypeUDP:
			udp := processUdpPacket(packet, ci)
			s.connections <- udp
			return
		}
	}
}

func (s *sensor) processConnections() {
	for connection := range s.connections {
		err := connection.analyze()
		if err != nil {
			log.Println(err)
		}
		logger.Log(*connection)
	}
}
