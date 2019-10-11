package gourmet

import (
	"log"

	"github.com/google/gopacket/afpacket"
)

func newAfpacketSensor(c *Config) (*afpacket.TPacket, error) {
	if c.Bpf != "" {
		log.Println("[*] Warning: filter option will not be applied when using afpacket sensor")
	}
	if c.Promiscuous == true {
		log.Println("[*] Warning: promiscuous mode not supported when using afpacket sensor")
	}
	tPacket, err := afpacket.NewTPacket(
		afpacket.OptFrameSize(c.SnapLen),
		afpacket.OptInterface(c.Interface))
	if err != nil {
		return nil, err
	}
	return tPacket, nil
}
