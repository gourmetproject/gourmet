package gourmet

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type interfaceType byte

const (
	afpacketType interfaceType = 1
	libpcapType  interfaceType = 3
)

type sensorMetadata struct {
	// The network interface that the sensor is capturing traffic
	NetworkInterface string
	// The IP address of the capturing network interface
	NetworkAddress []string
}

func getSensorMetadata(interfaceName string) *sensorMetadata {
	return &sensorMetadata{
		NetworkInterface: interfaceName,
		NetworkAddress:   getInterfaceAddresses(interfaceName),
	}
}

type sensor struct {
	source        gopacket.ZeroCopyPacketDataSource
	streamFactory *tcpStreamFactory
	connections   chan *Connection
}

// Start is the entry point for Gourmet
func Start(config *Config) {
	var err error
	var workingGraph analyzerGraph
	for k, v := range config.Analyzers {
		analyzerNode, err := createAnalyzerNode(k, v)
		if err != nil {
			log.Fatal(fmt.Errorf("unable to process analyzer config: %s", err))
		}
		workingGraph = append(workingGraph, analyzerNode)
	}
	err = resolveGraph(workingGraph)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to build dependency graph for analyzers: %s", err))
	}
	err = newAnalyzers(config.Analyzers, config.SkipUpdate)
	if err != nil {
		log.Fatal(err)
	}
	err = initLogger(config.LogFile, config.Interface)
	if err != nil {
		log.Fatal(err)
	}
	c := make(chan *Connection)
	s := &sensor{
		connections: c,
		streamFactory: &tcpStreamFactory{
			connections: c,
			connTimeout: config.ConnTimeout,
		},
	}
	err = s.getPacketSource(config)
	if err != nil {
		log.Fatal(err)
	}
	go s.processConnections()
	fmt.Printf("Gourmet is running and logging to %s. Press CTL+C to stop...", gLogger.fileName)
	fmt.Println()
	s.run()
}

func convertIfaceType(ifaceType string) (interfaceType, error) {
	if ifaceType == "libpcap" {
		return libpcapType, nil
	} else if ifaceType == "afpacket" {
		return afpacketType, nil
	} else {
		return 0, errors.New("invalid interface type. Must be libpcap or afpacket")
	}
}

func (s *sensor) getPacketSource(c *Config) (err error) {
	ifaceType, err := convertIfaceType(c.InterfaceType)
	if err != nil {
		return err
	}
	if ifaceType == afpacketType {
		s.source, err = newAfpacketSensor(c)
		if err != nil {
			return err
		}
	} else if ifaceType == libpcapType {
		s.source, err = newLibpcapSensor(c)
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
		p, ci, err := s.source.ZeroCopyReadPacketData()
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
			udp := processUDPPacket(packet, ci)
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
		gLogger.log(*connection)
	}
}
