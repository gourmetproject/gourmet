package gourmet

import (
    "github.com/google/gopacket/pfring"
)

type PfringSensor struct {
    options *SensorOptions
    ring    *pfring.Ring
}

func NewPfringSensor(opt *SensorOptions) (s *PfringSensor, err error) {
    ring, err := createPfring(opt.InterfaceName, opt.SnapLength, opt.IsPromiscuous)
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
    return &PfringSensor{
        opt,
        ring,
    }, nil
}

func createPfring(src string, snaplen uint32, promisc bool) (ring *pfring.Ring, err error) {
    if promisc {
        ring, err = pfring.NewRing(src, snaplen, pfring.FlagPromisc)
    } else {
        ring, err = pfring.NewRing(src, snaplen, 0)
    }
    return ring, err
}