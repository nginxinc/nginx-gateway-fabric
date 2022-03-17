package config

type server struct {
	ServerName string
	Locations  []location
}

type location struct {
	Path      string
	ProxyPass string
}
