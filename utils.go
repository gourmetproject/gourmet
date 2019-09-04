package gourmet

import (
    "github.com/google/gopacket"
    "strconv"
)

func getProtocol(transport gopacket.Flow) uint16 {
    a, _ := strconv.ParseUint(transport.Src().String(), 10, 16)
    b, _ := strconv.ParseUint(transport.Dst().String(), 10, 16)
    if a < b {
        return uint16(a)
    }
    return uint16(b)
}
