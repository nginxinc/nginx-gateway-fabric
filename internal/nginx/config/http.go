package config

type httpServers struct {
	Servers []server
}

type server struct {
	ServerName string
	Locations  []location
}

type location struct {
	Path         string
	ProxyPass    string
	HTTPMatchVar string
	Internal     bool
}
