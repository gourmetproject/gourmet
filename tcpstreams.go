package gourmet

import (
    "bytes"
    "fmt"
    "github.com/google/gopacket"
    "github.com/google/gopacket/layers"
    "github.com/google/gopacket/reassembly"
    "time"
)

type Protocol string
type TCPProtocolMap map[uint16]Protocol
type UDPProtocolMap map[uint16]Protocol
const (
    // TCP Protocols
    TlsProtocol    Protocol    = "tls"
    HttpProtocol   Protocol    = "http"
    DnsTcpProtocol Protocol    = "dns"
    // UDP Protocols
    DnsUdpProtocol Protocol    = "dns"
)
var (
    TcpProtocols = TCPProtocolMap {
        53:  DnsTcpProtocol,
        80:  HttpProtocol,
        443: TlsProtocol,
    }
    UdpProtocols = UDPProtocolMap {
        53: DnsUdpProtocol,
    }
)

type TcpStream struct {
    stream *tcpStream
}

func (t *TcpStream) Payload() []byte {
    return t.stream.payload.Bytes()
}

func (t *TcpStream) NetworkFlow() gopacket.Flow {
    return t.stream.net
}

func (t *TcpStream) TransportFlow() gopacket.Flow {
    return t.stream.transport
}

// tcpStream is an implementation of reassembly.Stream
type tcpStream struct {
    net, transport 	gopacket.Flow
    protocolType    Protocol
    payload         *bytes.Buffer
    done            chan bool
    packets         int
    payloadPackets  int
    tcpstate        *reassembly.TCPSimpleFSM
}

// Accept validates that the TCP stream is valid via reassembly.TCPSimpleFSM.CheckState before it
// passes it along to the assembler. A TCPSimpleFSM is assigned to each TCPStream upon creation by
// TcpStreamFactory.New, and updated each time we call this function. This ensures that each packet
// reassembled contains the proper TCP flags (SYN, ACK, etc.), depending on where it is sequentially
// in the transmission.
func (ts *tcpStream) Accept(tcp *layers.TCP, ci gopacket.CaptureInfo, dir reassembly.TCPFlowDirection, nextSeq reassembly.Sequence, start *bool, ac reassembly.AssemblerContext) bool {
    return ts.tcpstate.CheckState(tcp, dir)
}

// ReassembledSG implements the reassembly.Stream interface.
func (ts *tcpStream) ReassembledSG(sg reassembly.ScatterGather, ac reassembly.AssemblerContext) {
    length, _ := sg.Lengths()
    data := sg.Fetch(length)
    if length > 0 {
        ts.payload.Write(data)
    }
    ts.packets++
}

// ReassemblyComplete implements the reassembly.Stream interface.
func (ts *tcpStream) ReassemblyComplete(ac reassembly.AssemblerContext) bool {
    ts.done <- true
    return true
}

// tcpStreamFactory contains channels to consume tcp streams and stream pairs. It also implements
// the reassembly.StreamFactory interface. Each Sensor contains a tcpStreamFactory in order to
// easily consume packets, streams, and stream pairs.
type tcpStreamFactory struct {
    assembler *reassembly.Assembler
    streams chan *TcpStream
}

func (tsf *tcpStreamFactory) New(n, t gopacket.Flow, tcp *layers.TCP, ac reassembly.AssemblerContext) reassembly.Stream {
    protocol := getProtocol(t)
    s := &tcpStream {
        net:          n,
        transport:    t,
        payload:      new(bytes.Buffer),
        done:         make(chan bool),
        tcpstate:     reassembly.NewTCPSimpleFSM(reassembly.TCPSimpleFSMOptions{}),
        protocolType: TcpProtocols[protocol],
    }
    go func() {
        <- s.done
        // ignore empty streams flushed/closed early b/c of assembler timeout
        if s.packets > 0 {
            stream := &TcpStream{
                stream: s,
            }
            tsf.streams <- stream
        }
    }()
    return s
}

func (tsf *tcpStreamFactory) getTcpStreams(packets chan gopacket.Packet) {
    tsf.createAssembler()
    ticker := time.Tick(time.Minute)
    for {
        select {
        case packet := <- packets:
            var tcp *layers.TCP
            if packet.TransportLayer() == nil || packet.TransportLayer().LayerType() != layers.LayerTypeTCP {
                continue
            }
            tcp = packet.TransportLayer().(*layers.TCP)
            tsf.assembler.Assemble(packet.NetworkLayer().NetworkFlow(), tcp)
        case <- ticker:
            fmt.Println(tsf.assembler.Dump())
            flushed, closed := tsf.assembler.FlushCloseOlderThan(time.Now().Add(time.Minute * -4))
            fmt.Println(flushed, closed)
        }
    }
}

func (tsf *tcpStreamFactory) createAssembler() {
    streamPool := reassembly.NewStreamPool(tsf)
    tsf.assembler = reassembly.NewAssembler(streamPool)
}