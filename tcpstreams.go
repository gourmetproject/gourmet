package gourmet

import (
    "bytes"
    "github.com/google/gopacket"
    "github.com/google/gopacket/layers"
    "github.com/google/gopacket/reassembly"
    "time"
)

type protocol string
type TCPProtocolMap map[uint16]protocol
type UDPProtocolMap map[uint16]protocol
const (
    // TCP Protocols
    TlsProtocol    protocol = "tls"
    HttpProtocol   protocol = "http"
    DnsTcpProtocol protocol = "dns"
    // UDP Protocols
    DnsUdpProtocol protocol = "dns"
)
var (
    tcpProtocols = TCPProtocolMap {
        53:  DnsTcpProtocol,
        80:  HttpProtocol,
        443: TlsProtocol,
    }
    udpProtocols = UDPProtocolMap {
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

func (ts *tcpStream) Accept(tcp *layers.TCP, ci gopacket.CaptureInfo, dir reassembly.TCPFlowDirection, nextSeq reassembly.Sequence, start *bool, ac reassembly.AssemblerContext) bool {
    return true
}

func (ts *tcpStream) ReassembledSG(sg reassembly.ScatterGather, ac reassembly.AssemblerContext) {
    length, _ := sg.Lengths()
    data := sg.Fetch(length)
    if length > 0 {
        ts.payload.Write(data)
    }
    ts.packets++
}

func (ts *tcpStream) ReassemblyComplete(ac reassembly.AssemblerContext) bool {
    ts.done <- true
    return true
}

// tcpStreamFactory contains channels to consume tcp streams and stream pairs. It also implements
// the reassembly.StreamFactory interface. Each Sensor contains a tcpStreamFactory in order to
// easily consume packets, streams, and stream pairs.
type tcpStreamFactory struct {
    assembler *reassembly.Assembler
    streams   chan *TcpStream
    ticker    *time.Ticker
}

func (tsf *tcpStreamFactory) New(n, t gopacket.Flow, tcp *layers.TCP, ac reassembly.AssemblerContext) reassembly.Stream {
    protocol := getTcpProtocol(t)
    s := &tcpStream {
        net:          n,
        transport:    t,
        payload:      new(bytes.Buffer),
        done:         make(chan bool),
        tcpstate:     reassembly.NewTCPSimpleFSM(reassembly.TCPSimpleFSMOptions{}),
        protocolType: protocol,
    }
    go func() {
        // wait for reassembly to be done
        <- s.done
        // ignore empty streams
        if s.packets > 0 {
            stream := &TcpStream{
                stream: s,
            }
            tsf.streams <- stream
        }
    }()
    return s
}

func (tsf *tcpStreamFactory) assemblePacket(netFlow gopacket.Flow, tcp *layers.TCP) {
    tsf.assembler.Assemble(netFlow, tcp)
    select {
    case <- tsf.ticker.C:
        tsf.assembler.FlushCloseOlderThan(time.Now().Add(time.Second * -40))
    default:
    }
}

func (tsf *tcpStreamFactory) createAssembler() {
    streamPool := reassembly.NewStreamPool(tsf)
    tsf.assembler = reassembly.NewAssembler(streamPool)
}