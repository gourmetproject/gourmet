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
		streamFactory: &TcpStreamFactory{
			streams: make(chan *TcpStream),
		},
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
	go s.processStreams()
	for {
		p, _, err := s.source.ReadPacketData()
		if err != nil {
			log.Println(err)
			continue
		}
		packet := gopacket.NewPacket(p, layers.LayerTypeEthernet, gopacket.Default)
		go s.processNewPacket(packet)
	}
}

func (s *Sensor) processNewPacket(packet gopacket.Packet) {
	if packet.TransportLayer() != nil {
		switch packet.TransportLayer().LayerType() {
		case layers.LayerTypeTCP:
			s.streamFactory.newPacket(packet.NetworkLayer().NetworkFlow(), packet.TransportLayer().(*layers.TCP))
		case layers.LayerTypeUDP:
			// TODO: UDP packet
		}
	}
}

func (s *Sensor) processStreams() {
	for stream := range s.streamFactory.streams {
		go processStream(stream)
	}
}

func processStream(stream *TcpStream) {
	c := newTcpConnection(stream)
	for i, analyzer := range analyzers {
		if analyzer.inUse && analyzer.Filter(c) {
			fmt.Println(i)
			result, err := analyzer.Analyze(c)
			if err != nil {
				log.Println(err)
			}
			c.Analyzers[result.Key()] = result
		}
	}
	logger.Log(c)
}
