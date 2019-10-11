package gourmet

import (
	"bytes"

	"github.com/google/gopacket"
)

func processUDPPacket(packet gopacket.Packet, ci gopacket.CaptureInfo) *Connection {
	srcPort, dstPort := processPorts(packet.TransportLayer().TransportFlow())
	return &Connection{
		Timestamp:       ci.Timestamp,
		UID:             packet.NetworkLayer().NetworkFlow().FastHash() + packet.TransportLayer().TransportFlow().FastHash(),
		SourceIP:        packet.NetworkLayer().NetworkFlow().Src().String(),
		SourcePort:      srcPort,
		DestinationIP:   packet.NetworkLayer().NetworkFlow().Dst().String(),
		DestinationPort: dstPort,
		TransportType:   "udp",
		Payload:         bytes.NewBuffer(packet.TransportLayer().LayerPayload()),
		Analyzers:       make(map[string]interface{}),
	}
}
