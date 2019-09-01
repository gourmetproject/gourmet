package gourmet

import "github.com/google/gopacket/afpacket"

type AfpacketSensor struct {
    options *SensorOptions
    tPacket *afpacket.TPacket
}