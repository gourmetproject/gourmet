package gourmet

import "github.com/google/gopacket/layers"

type HttpStream struct {
    tcpStream *TcpStream
}

type HttpStreamFactory struct {}

func (hsf *HttpStreamFactory) New(ts *TcpStream) HttpStream {
    stream := HttpStream{
        tcpStream: ts,
    }
    return stream
}

func (hs *HttpStream) ApplicationProtocol() (protocolName string) {
    for k, protocol := range TcpProtocols {
        if hs.tcpStream.ProtocolType == protocol {
            protocolName = layers.TCPPortNames[layers.TCPPort(k)]
            break
        }
    }
    return protocolName
}