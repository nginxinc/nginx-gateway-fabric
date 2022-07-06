package config

type httpServers struct {
	Servers []server
}

type server struct {
	IsDefault  bool
	ServerName string
	SSL        *ssl
	Locations  []location
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
