package gourmet

import "github.com/google/gopacket/pcap"

type LibpcapSensor struct {
    options *SensorOptions
    handle *pcap.Handle
}