package gourmet

type protocol string

const (
	DnsProtocol  protocol = "dns"
	HttpProtocol protocol = "http"
	TlsProtocol  protocol = "tls"
)

var (
	protocolMap = map[uint16]protocol {
		53:  DnsProtocol,
		80:  HttpProtocol,
		443: TlsProtocol,
	}
)
