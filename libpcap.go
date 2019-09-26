package gourmet

import (
	"github.com/google/gopacket/pcap"
)

func newLibpcapSensor(opt *SensorOptions) (*pcap.Handle, error) {
	err := initOptions(opt)
	if err != nil {
		return nil, err
	}
	var handle *pcap.Handle
	handle, err = pcap.OpenLive(opt.InterfaceName, int32(opt.SnapLen), opt.IsPromiscuous, pcap.BlockForever)
	if err != nil {
		return nil, err
	}
	err = handle.SetBPFFilter(opt.Bpf)
	if err != nil {
		return nil, err
	}
	return handle, nil
}
