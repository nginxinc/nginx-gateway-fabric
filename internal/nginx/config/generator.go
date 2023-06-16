package config

import (
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/file"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/dataplane"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Generator

// Generator generates NGINX configuration.
// This interface is used for testing purposes only.
type Generator interface {
	// Generate generates NGINX configuration from internal representation.
	Generate(configuration dataplane.Configuration) []file.File
}

// GeneratorImpl is an implementation of Generator.
type GeneratorImpl struct{}

// NewGeneratorImpl creates a new GeneratorImpl.
func NewGeneratorImpl() GeneratorImpl {
	return GeneratorImpl{}
}

// executeFunc is a function that generates NGINX configuration from internal representation.
type executeFunc func(configuration dataplane.Configuration) []byte

// Generate generates NGINX configuration from internal representation.
// It is the responsibility of the caller to validate the configuration before calling this function.
// In case of invalid configuration, NGINX will fail to reload or could be configured with malicious configuration.
// To validate, use the validators from the validation package.
func (g GeneratorImpl) Generate(conf dataplane.Configuration) []file.File {
	files := make([]file.File, 0, len(conf.TLSCerts)+1)

	// certs
	for id, cert := range conf.TLSCerts {
		contents := make([]byte, 0, len(cert.Cert)+len(cert.Key)+1)
		contents = append(contents, cert.Cert...)
		contents = append(contents, '\n')
		contents = append(contents, cert.Key...)

		files = append(files, file.File{
			Content:     contents,
			Path:        generateTLSCertPath(id),
			Permissions: 0600,
		})
	}

	var generated []byte
	for _, execute := range getExecuteFuncs() {
		generated = append(generated, execute(conf)...)
	}

	files = append(files, file.File{
		Content:     generated,
		Path:        "/etc/nginx/conf.d/http.conf",
		Permissions: 0644,
	})

	return files
}

func generateTLSCertPath(id dataplane.TLSCertID) string {
	return "/etc/nginx/secrets/" + string(id) + ".pem"
}

func getExecuteFuncs() []executeFunc {
	return []executeFunc{
		executeUpstreams,
		executeSplitClients,
		executeServers,
	}
}
