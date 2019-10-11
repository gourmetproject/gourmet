package gourmet

import (
	"bytes"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/reassembly"
)

type tcpStream struct {
	net, transport gopacket.Flow
	payload        *bytes.Buffer
	startTime      time.Time
	duration       time.Duration
	tcpState       *reassembly.TCPSimpleFSM
	done           chan bool
	packets        int
	payloadPackets int
}

func newConnectionFromTCP(ts *tcpStream) (c *Connection) {
	srcPort, dstPort := processPorts(ts.transport)
	return &Connection{
		Timestamp:       ts.startTime,
		UID:             ts.net.FastHash() + ts.transport.FastHash(),
		SourceIP:        ts.net.Src().String(),
		SourcePort:      srcPort,
		DestinationIP:   ts.net.Dst().String(),
		DestinationPort: dstPort,
		TransportType:   "tcp",
		Duration:        ts.duration.Seconds(),
		State:           ts.tcpState.String(),
		Payload:         ts.payload,
		Analyzers:       make(map[string]interface{}),
	}
}

func (ts *tcpStream) Accept(tcp *layers.TCP, ci gopacket.CaptureInfo, dir reassembly.TCPFlowDirection, nextSeq reassembly.Sequence, start *bool, ac reassembly.AssemblerContext) bool {
	tempDuration := ci.Timestamp.Sub(ts.startTime)
	if tempDuration.Seconds() > ts.duration.Seconds() {
		ts.duration = tempDuration
	}
	ts.tcpState.CheckState(tcp, dir)
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
	return false
}

// tcpStreamFactory contains channels to consume tcp streams and stream pairs. It also implements
// the reassembly.StreamFactory interface. Each Sensor contains a tcpStreamFactory in order to
// easily consume packets, streams, and stream pairs.
type tcpStreamFactory struct {
	assembler      *reassembly.Assembler
	assemblerMutex sync.Mutex
	connTimeout    int
	ticker         *time.Ticker
	connections    chan *Connection
}

func (tsf *tcpStreamFactory) New(n, t gopacket.Flow, tcp *layers.TCP, ac reassembly.AssemblerContext) reassembly.Stream {
	ts := &tcpStream{
		net:       n,
		transport: t,
		payload:   new(bytes.Buffer),
		startTime: ac.GetCaptureInfo().Timestamp,
		tcpState:  reassembly.NewTCPSimpleFSM(reassembly.TCPSimpleFSMOptions{}),
		done:      make(chan bool),
	}
	go func() {
		// wait for reassembly to be done
		<-ts.done
		// ignore empty streams
		if ts.packets > 0 {
			c := newConnectionFromTCP(ts)
			tsf.connections <- c
		}
	}()
	return ts
}

func (tsf *tcpStreamFactory) newPacket(netFlow gopacket.Flow, tcp *layers.TCP) {
	select {
	case <-tsf.ticker.C:
		tsf.assemblerMutex.Lock()
		tsf.assembler.FlushCloseOlderThan(time.Now().Add(time.Second * time.Duration(-1*tsf.connTimeout)))
		tsf.assemblerMutex.Unlock()
	default:
		// pass through
	}
	tsf.assemblePacket(netFlow, tcp)
}

func (tsf *tcpStreamFactory) assemblePacket(netFlow gopacket.Flow, tcp *layers.TCP) {
	tsf.assemblerMutex.Lock()
	tsf.assembler.Assemble(netFlow, tcp)
	tsf.assemblerMutex.Unlock()
}

func (tsf *tcpStreamFactory) createAssembler() {
	streamPool := reassembly.NewStreamPool(tsf)
	tsf.assembler = reassembly.NewAssembler(streamPool)
}
