package gourmet

import (
    "fmt"
    "github.com/google/gopacket"
    "github.com/google/gopacket/layers"
    "github.com/google/gopacket/reassembly"
    "io"
    "strconv"
    "time"
)

type Protocol byte
type TCPProtocolMap map[uint16]Protocol
type UDPProtocolMap map[uint16]Protocol
const (
    // TCP Protocols
    TlsProtocol    Protocol    = 1
    HttpProtocol   Protocol    = 2
    DnsTcpProtocol Protocol    = 3
    // UDP Protocols
    DnsUdpProtocol Protocol    = 4
)
var (
    TcpProtocols = TCPProtocolMap {
        53: DnsTcpProtocol,
        80: HttpProtocol,
        443: TlsProtocol,
    }
    UdpProtocols = UDPProtocolMap {
        53: DnsUdpProtocol,
    }
)

// A TcpStreamPair contains the bidirectional data flow for a tcp connection. A stream is identified
// as the Request stream if the source port is larger than the destination port. Naturally, the
// Response stream is the response to this request.
//
// If two hashes generated from Transport.FastHash() match, then we know the
type TcpStreamPair struct {
    Request  *TcpStream
    Response *TcpStream
}

type tcpStreamFactory struct {
    sensor *Sensor
}

// TcpStream is an implementation of reassembly.Stream. It is also an implementation of an io.Reader
// in order to easily consume TCP payloads.
type TcpStream struct {
    Net, Transport 	gopacket.Flow
    payload         []byte
    byteChannel     chan []byte
    Time            time.Time
    tcpstate        *reassembly.TCPSimpleFSM
    ProtocolType    Protocol
}

func (ts *TcpStream) Read(p []byte) (int, error) {
    ok := true
    for ok && len(ts.payload) == 0 {
        ts.payload, ok = <-ts.byteChannel
    }
    if !ok || len(ts.payload) == 0 {
        return 0, io.EOF
    }

    l := copy(p, ts.payload)
    ts.payload = ts.payload[l:]
    return l, nil
}

// Accept validates that the TCP stream is valid via reassembly.TCPSimpleFSM.CheckState(tcp, dir)
func (ts *TcpStream) Accept(
    tcp *layers.TCP,
    ci gopacket.CaptureInfo,
    dir reassembly.TCPFlowDirection,
    nextSeq reassembly.Sequence,
    start *bool,
    ac reassembly.AssemblerContext,
) bool {
    return ts.tcpstate.CheckState(tcp, dir)
}

func (ts *TcpStream) ReassembledSG(sg reassembly.ScatterGather, ac reassembly.AssemblerContext) {
    length, _ := sg.Lengths()
    if length > 0 {
        fmt.Println(length)
        data := sg.Fetch(length)
        ts.byteChannel <- data
    }
}

func (ts *TcpStream) ReassemblyComplete(ac reassembly.AssemblerContext) bool {
    close(ts.byteChannel)
    return true
}

func (tsf *tcpStreamFactory) New(n, t gopacket.Flow, tcp *layers.TCP, ac reassembly.AssemblerContext) reassembly.Stream {
    stream := &TcpStream {
        Net:          n,
        Transport:    t,
        Time:         time.Now(),
        byteChannel:  make(chan []byte),
        tcpstate:     reassembly.NewTCPSimpleFSM(reassembly.TCPSimpleFSMOptions{}),
        ProtocolType: TcpProtocols[getProtocol(t)],
    }
    go func() {
        tsf.sensor.streams <- stream
    }()
    return stream
}

func createAssembler(factory *tcpStreamFactory) *reassembly.Assembler {
    streamPool := reassembly.NewStreamPool(factory)
    assembler := reassembly.NewAssembler(streamPool)
    return assembler
}

func getProtocol(transport gopacket.Flow) uint16 {
    a, _ := strconv.ParseUint(transport.Src().String(), 10, 16)
    b, _ := strconv.ParseUint(transport.Dst().String(), 10, 16)
    if a < b {
        return uint16(a)
    }
    return uint16(b)

}