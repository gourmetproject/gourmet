package gourmet

import (
	"fmt"
	"github.com/google/gopacket"
)

type DnsLog struct {
	id        int64
	queryType string
	query     string
}

func processDnsStream(tcp *TcpStream) {
	fmt.Println("dns stream found")
}

func processDnsPacket(packet gopacket.Packet) {
	fmt.Println("dns packet found")
}
