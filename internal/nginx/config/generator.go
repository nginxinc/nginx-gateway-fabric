package config

import (
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Generator

// Generator generates NGINX configuration.
// This interface is used for testing purposes only.
type Generator interface {
	// Generate generates NGINX configuration from internal representation.
	Generate(configuration state.Configuration) []byte
}

// GeneratorImpl is an implementation of Generator.
type GeneratorImpl struct{}

// NewGeneratorImpl creates a new GeneratorImpl.
func NewGeneratorImpl() GeneratorImpl {
	return GeneratorImpl{}
}

// executeFunc is a function that generates NGINX configuration from internal representation.
type executeFunc func(configuration state.Configuration) []byte

func (g GeneratorImpl) Generate(conf state.Configuration) []byte {
	var generated []byte
	for _, execute := range getExecuteFuncs() {
		generated = append(generated, execute(conf)...)
	}

	return generated
}

func getExecuteFuncs() []executeFunc {
	return []executeFunc{
		executeUpstreams,
		executeSplitClients,
		executeServers,
	}
}
