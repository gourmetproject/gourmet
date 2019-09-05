package gourmet

import (
    "github.com/google/gopacket/pcap"
    "time"
)

func newLibpcapSensor(opt *SensorOptions) (*pcap.Handle, error) {
    err := initOptions(opt)
    if err != nil {
        return nil, err
    }
    var handle *pcap.Handle
    if opt.Timeout == 0 {
        handle, err = pcap.OpenLive(
            opt.InterfaceName, int32(opt.SnapLen), opt.IsPromiscuous, pcap.BlockForever)
    } else {
        handle, err = pcap.OpenLive(
            opt.InterfaceName,
            int32(opt.SnapLen),
            opt.IsPromiscuous,
            time.Duration(opt.Timeout) * time.Second)
    }
    if err != nil {
        return nil, err
    }
    err = handle.SetBPFFilter(opt.Filter)
    if err != nil {
        return nil, err
    }
    return handle, nil
}
