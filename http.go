package gourmet

type HttpStream struct {
    tcpStream *TcpStream
}

func (hs *HttpStream) GetProtocol() Protocol {
    return hs.tcpStream.stream.protocolType
}