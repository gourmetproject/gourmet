package gourmet

import (
    "github.com/google/gopacket"
    "github.com/google/gopacket/afpacket"
    "github.com/google/gopacket/layers"
    "log"
)

func newAfpacketSensor(opt *SensorOptions) (src *gopacket.PacketSource, err error) {
    err = initOptions(opt)
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
    return gopacket.NewPacketSource(tPacket, layers.LayerTypeEthernet), nil
}