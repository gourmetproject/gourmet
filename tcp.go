package gourmet

import (
	"bytes"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/reassembly"
	"sync"
	"time"
)

// tcpStream is an implementation of reassembly.Stream
type TcpStream struct {
	net, transport gopacket.Flow
	protocolType   protocol
	payload        *bytes.Buffer
	startTime      time.Time
	duration       time.Duration
	tcpState       *reassembly.TCPSimpleFSM
	done           chan bool
	packets        int
	payloadPackets int
}

func (ts *TcpStream) Accept(tcp *layers.TCP, ci gopacket.CaptureInfo, dir reassembly.TCPFlowDirection, nextSeq reassembly.Sequence, start *bool, ac reassembly.AssemblerContext) bool {
	tempDuration := ci.Timestamp.Sub(ts.startTime)
	if tempDuration.Seconds() > ts.duration.Seconds() {
		ts.duration = tempDuration
	}
	ts.tcpState.CheckState(tcp, dir)
	return true
}

func (ts *TcpStream) ReassembledSG(sg reassembly.ScatterGather, ac reassembly.AssemblerContext) {
	length, _ := sg.Lengths()
	data := sg.Fetch(length)
	if length > 0 {
		ts.payload.Write(data)
	}
	ts.packets++
}

func (ts *TcpStream) ReassemblyComplete(ac reassembly.AssemblerContext) bool {
	ts.done <- true
	return false
}

// tcpStreamFactory contains channels to consume tcp streams and stream pairs. It also implements
// the reassembly.StreamFactory interface. Each Sensor contains a tcpStreamFactory in order to
// easily consume packets, streams, and stream pairs.
type TcpStreamFactory struct {
	assembler      *reassembly.Assembler
	assemblerMutex sync.Mutex
	ticker         *time.Ticker
	streams        chan *TcpStream
}

func (tsf *TcpStreamFactory) New(n, t gopacket.Flow, tcp *layers.TCP, ac reassembly.AssemblerContext) reassembly.Stream {
	protocol := getTcpProtocol(t)
	ts := &TcpStream{
		net:          n,
		transport:    t,
		payload:      new(bytes.Buffer),
		startTime:    ac.GetCaptureInfo().Timestamp,
		tcpState:     reassembly.NewTCPSimpleFSM(reassembly.TCPSimpleFSMOptions{}),
		done:         make(chan bool),
		protocolType: protocol,
	}
	go func() {
		// wait for reassembly to be done
		<-ts.done
		// ignore empty streams
		if ts.packets > 0 {
			tsf.streams <- ts
		}
	}()
	return ts
}

func (tsf *TcpStreamFactory) newPacket(netFlow gopacket.Flow, tcp *layers.TCP) {
	select {
	case <-tsf.ticker.C:
		fmt.Println("flushing")
		tsf.assembler.FlushCloseOlderThan(time.Now().Add(time.Second * -40))
	default:
		// pass through
	}
	tsf.assemblePacket(netFlow, tcp)
}

func (tsf *TcpStreamFactory) assemblePacket(netFlow gopacket.Flow, tcp *layers.TCP) {
	tsf.assemblerMutex.Lock()
	tsf.assembler.Assemble(netFlow, tcp)
	tsf.assemblerMutex.Unlock()
}

func (tsf *TcpStreamFactory) createAssembler() {
	streamPool := reassembly.NewStreamPool(tsf)
	tsf.assembler = reassembly.NewAssembler(streamPool)
}
