package gourmet

import (
	"github.com/google/gopacket/pcap"
)

func newLibpcapSensor(c *Config) (*pcap.Handle, error) {
	var handle *pcap.Handle
	handle, err := pcap.OpenLive(c.Interface, int32(c.SnapLen), c.Promiscuous, pcap.BlockForever)
	if err != nil {
		return nil, err
	}
	err = handle.SetBPFFilter(c.Bpf)
	if err != nil {
		return nil, err
	}
	return handle, nil
}
