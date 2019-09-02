package gourmet

import (
    "github.com/google/gopacket"
    "github.com/google/gopacket/layers"
    "github.com/google/gopacket/pfring"
)

func newPfringSensor(opt *SensorOptions) (src *gopacket.PacketSource, err error) {
    err = initOptions(opt)
    if err != nil {
        return nil, err
    }
    ring, err := createPfring(opt)
    if err != nil {
        return nil, err
    }
    if opt.Filter != "" {
        err = ring.SetBPFFilter(opt.Filter)
        if err != nil {
            return nil, err
        }
    }
    err = ring.Enable()
    if err != nil {
        return nil, err
    }
    return gopacket.NewPacketSource(ring, layers.LayerTypeEthernet), nil
}

func createPfring(opt *SensorOptions) (ring *pfring.Ring, err error) {
    if opt.IsPromiscuous {
        ring, err = pfring.NewRing(opt.InterfaceName, opt.SnapLen, pfring.FlagPromisc)
    } else {
        ring, err = pfring.NewRing(opt.InterfaceName, opt.SnapLen, 0)
    }
    return ring, err
}