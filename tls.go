package gourmet

import (
	"fmt"
	"github.com/google/gopacket"
)

func processTlsStream(tcp *TcpStream) {
	fmt.Println("TLS found")
}

func processTlsPacket(packet gopacket.Packet) {

}
