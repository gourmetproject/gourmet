package gourmet

type ApplicationStream interface {
    ApplicationProtocol() string
}

type ApplicationStreamFactory interface {
    New(tcpStream *TcpStream) ApplicationStream
}
