package config

import (
	"bytes"
	"fmt"
	"os"

	"github.com/nginx/agent/sdk/v2/proto"
	"github.com/nginx/agent/sdk/v2/zip"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/config"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/observer"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/dataplane"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/secrets"
)

const (
	secretsFileMode   = 0o600
	confPrefix        = "/etc/nginx"
	secretsPrefix     = "/etc/nginx/secrets" //nolint:gosec
	nginxConfFilePath = "nginx.conf"
	httpConfFilePath  = "conf.d/http.conf"
)

// NginxConfig is an intermediate object that contains nginx configuration in a form that agent expects.
// We convert the dataplane configuration to NginxConfig in the config store, so we only need to do it once
// per configuration change. The NginxConfig is then used by the agent to generate the nginx configuration payload.
type NginxConfig struct {
	version     string
	config      *proto.ZippedFile
	aux         *proto.ZippedFile
	directories []*proto.Directory
}

func (n *NginxConfig) GetVersion() string {
	return n.version
}

type directory struct {
	prefix string
	files  []file
}

type file struct {
	path     string
	contents []byte
	mode     os.FileMode
}

// NginxConfigAdapter adapts the dataplane.Configuration to NginxConfig.
type NginxConfigAdapter struct {
	generator        config.Generator
	secretRequestMgr secrets.RequestManager
}

// NewNginxConfigAdapter creates a new NginxConfigAdapter.
func NewNginxConfigAdapter(
	generator config.Generator,
	secretRequestMgr secrets.RequestManager,
) *NginxConfigAdapter {
	return &NginxConfigAdapter{
		generator:        generator,
		secretRequestMgr: secretRequestMgr,
	}
}

// VersionedConfig adapts the dataplane.Configuration to NginxConfig.
// It uses the config.Generator to generate the nginx.conf and http.conf files,
// and the secrets.RequestManager to generate the secret files.
// Implements VersionedConfigAdapter.
func (u *NginxConfigAdapter) VersionedConfig(cfg dataplane.Configuration) (observer.VersionedConfig, error) {
	confDirectory := u.generateConfigDirectory(cfg)
	auxDirectory := u.generateAuxConfigDirectory()

	directories := []*proto.Directory{
		convertToProtoDirectory(confDirectory),
		convertToProtoDirectory(auxDirectory),
	}

	zconfig, err := generateZippedFile(confDirectory)
	if err != nil {
		return nil, err
	}

	zaux, err := generateZippedFile(auxDirectory)
	if err != nil {
		return nil, err
	}

	return &NginxConfig{
		version:     fmt.Sprintf("%d", cfg.Version),
		config:      zconfig,
		aux:         zaux,
		directories: directories,
	}, nil
}

func (u *NginxConfigAdapter) generateConfigDirectory(cfg dataplane.Configuration) directory {
	return directory{
		prefix: confPrefix,
		files: []file{
			{
				path:     nginxConfFilePath,
				mode:     secretsFileMode,
				contents: u.generator.GenerateMainConf(cfg.Version),
			},
			{
				path:     httpConfFilePath,
				mode:     secretsFileMode,
				contents: u.generator.GenerateHTTPConf(cfg),
			},
		},
	}
}

func (u *NginxConfigAdapter) generateAuxConfigDirectory() directory {
	secretFiles := u.secretRequestMgr.GetAndResetRequestedSecrets()

	files := make([]file, 0, len(secretFiles))
	for _, secret := range secretFiles {
		files = append(files, file{
			path:     secret.Name,
			mode:     secretsFileMode,
			contents: secret.Contents,
		})
	}

	return directory{
		prefix: secretsPrefix,
		files:  files,
	}
}

func convertToProtoDirectory(d directory) *proto.Directory {
	files := make([]*proto.File, len(d.files))

	for idx, f := range d.files {
		files[idx] = &proto.File{
			Name: f.path,
		}
	}

	return &proto.Directory{
		Name:  d.prefix,
		Files: files,
	}
}

func generateZippedFile(dir directory) (*proto.ZippedFile, error) {
	w, err := zip.NewWriter(dir.prefix)
	if err != nil {
		return nil, err
	}

	for _, f := range dir.files {
		if err := w.Add(f.path, f.mode, bytes.NewBuffer(f.contents)); err != nil {
			return nil, err
		}
	}

	contents, prefix, checksum, err := w.Payloads()
	if err != nil {
		return nil, err
	}

	zipFile := &proto.ZippedFile{
		Contents:      contents,
		Checksum:      checksum,
		RootDirectory: prefix,
	}

	return zipFile, nil
}
