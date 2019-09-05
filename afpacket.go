package gourmet

import (
    "github.com/google/gopacket/afpacket"
    "log"
)

func newAfpacketSensor(opt *SensorOptions) (*afpacket.TPacket, error) {
    err := initOptions(opt)
    if err != nil {
        return nil, err
    }
    if opt.Filter != "" {
        log.Println("Warning: filter option will not be applied when using afpacket sensor")
    }
    tPacket, err := afpacket.NewTPacket(
        afpacket.OptFrameSize(opt.SnapLen),
        afpacket.OptInterface(opt.InterfaceName))
    if err != nil {
        return nil, err
    }
    return tPacket, nil
}