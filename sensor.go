package gourmet

import (
	"errors"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"log"
	"runtime"
	"time"
)

type interfaceType byte

const (
	// packet capture types
	AfpacketType interfaceType = 1
	PfringType   interfaceType = 2
	LibpcapType  interfaceType = 3
)
/*
SENSOR METADATA
 */
type SensorMetadata struct {
	Cores int
	NetworkInterface string
	NetworkAddress   []string
}

func getSensorMetadata(interfaceName string) *SensorMetadata{
	return &SensorMetadata{
		Cores:            runtime.NumCPU(),
		NetworkInterface: interfaceName,
		NetworkAddress:   getInterfaceAddresses(interfaceName),
	}
}

/*
SENSOR OPTIONS
 */
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

/*
SENSOR
 */
type Sensor struct {
	source        gopacket.PacketDataSource
	streamFactory *tcpStreamFactory
	connections   chan *Connection
}

func Start(options *SensorOptions) {
	var err error
	logger, err = newLogger("gourmet.log", options.InterfaceName)
	if err != nil {
		log.Fatal(err)
	}
	c := make(chan *Connection)
	s := &Sensor{
		connections: c,
		streamFactory: &tcpStreamFactory{
			connections: c,
		},
	}
	err = s.getPacketSource(options)
	if err != nil {
		log.Fatal(err)
	}
	go s.processConnections()
	s.run()
}

func (s *Sensor) getPacketSource(options *SensorOptions) (err error) {
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

func (s *Sensor) run() {
	s.streamFactory.createAssembler()
	s.streamFactory.ticker = time.NewTicker(time.Second * 10)
	for {
		p, ci, err := s.source.ReadPacketData()
		if err != nil {
			log.Println(err)
			continue
		}
		packet := gopacket.NewPacket(p, layers.LayerTypeEthernet, gopacket.Default)
		go s.processNewPacket(packet, ci)
	}
}

func (s *Sensor) processNewPacket(packet gopacket.Packet, ci gopacket.CaptureInfo) {
	if packet.TransportLayer() != nil {
		switch packet.TransportLayer().LayerType() {
		case layers.LayerTypeTCP:
			s.streamFactory.newPacket(packet.NetworkLayer().NetworkFlow(), packet.TransportLayer().(*layers.TCP))
		case layers.LayerTypeUDP:
			udp := processUdpPacket(packet, ci)
			s.connections <- udp
		}
	}
}

func (s *Sensor) processConnections() {
	for connection := range s.connections {
		err := connection.analyze()
		if err != nil {
			log.Println(err)
		}
		logger.Log(*connection)
	}
}
