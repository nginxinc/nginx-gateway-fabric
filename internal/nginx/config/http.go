package config

type httpServers struct {
	Servers []server
}

type server struct {
	SSL           *ssl
	ServerName    string
	Locations     []location
	IsDefaultHTTP bool
	IsDefaultSSL  bool
}

type location struct {
	Return       *returnVal
	Path         string
	ProxyPass    string
	HTTPMatchVar string
	Internal     bool
}

type returnVal struct {
	Code statusCode
}

type ssl struct {
	Certificate    string
	CertificateKey string
}

type statusCode int

const statusNotFound statusCode = 404
