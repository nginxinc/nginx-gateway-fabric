package config

import (
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/dataplane"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Generator

// Generator generates NGINX configuration.
// This interface is used for testing purposes only.
type Generator interface {
	// GenerateHTTPConf generates NGINX HTTP configuration from internal representation.
	GenerateHTTPConf(configuration dataplane.Configuration) []byte
	// GenerateMainConf generates the main nginx.conf file.
	GenerateMainConf(configGeneration int) []byte
}

// GeneratorImpl is an implementation of Generator.
type GeneratorImpl struct{}

// NewGeneratorImpl creates a new GeneratorImpl.
func NewGeneratorImpl() GeneratorImpl {
	return GeneratorImpl{}
}

// executeFunc is a function that generates NGINX configuration from internal representation.
type executeFunc func(configuration dataplane.Configuration) []byte

func getExecuteFuncs() []executeFunc {
	return []executeFunc{
		executeUpstreams,
		executeSplitClients,
		executeServers,
	}
}

// GenerateHTTPConf generates NGINX HTTP configuration from internal representation.
func (g GeneratorImpl) GenerateHTTPConf(conf dataplane.Configuration) []byte {
	var generated []byte
	for _, execute := range getExecuteFuncs() {
		generated = append(generated, execute(conf)...)
	}

	return generated
}

// GenerateMainConf generates the main nginx.conf file using the given configGeneration.
func (g GeneratorImpl) GenerateMainConf(configGeneration int) []byte {
	return executeNginxConf(configGeneration)
}
