package config

type http struct {
	Servers   []server
	Upstreams []upstream
}

type server struct {
	IsDefaultHTTP bool
	IsDefaultSSL  bool
	ServerName    string
	SSL           *ssl
	Locations     []location
}

type location struct {
	Path         string
	ProxyPass    string
	HTTPMatchVar string
	Internal     bool
}

type ssl struct {
	Certificate    string
	CertificateKey string
}

type upstream struct {
	Name    string
	Servers []upstreamServer
}

type upstreamServer struct {
	Address string
}
