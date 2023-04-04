package agent

import (
	"bytes"
	"fmt"
	"os"

	"github.com/nginx/agent/sdk/v2/proto"
	"github.com/nginx/agent/sdk/v2/zip"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/config"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/dataplane"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/secrets"
)

const (
	// TODO: do we need another file mode for config files?
	secretsFileMode   = 0o600
	confPrefix        = "/etc/nginx"
	secretsPrefix     = "/etc/nginx/secrets" //nolint:gosec
	nginxConfFilePath = "nginx.conf"
	httpConfFilePath  = "/conf.d/http.conf"
)

type directory struct {
	prefix string
	files  []file
}

type file struct {
	path     string
	contents []byte
	mode     os.FileMode
}

// NginxConfigBuilder builds NginxConfig from the dataplane configuration.
type NginxConfigBuilder struct {
	generator    config.Generator
	secretMemMgr secrets.SecretDiskMemoryManager
}

// NewNginxConfigBuilder creates a new NginxConfigBuilder.
func NewNginxConfigBuilder(
	generator config.Generator,
	secretMemMgr secrets.SecretDiskMemoryManager,
) *NginxConfigBuilder {
	return &NginxConfigBuilder{
		generator:    generator,
		secretMemMgr: secretMemMgr,
	}
}

// Build builds NginxConfig from the dataplane configuration.
// It generates the nginx configuration files using the config.Generator and the
// secrets files using the secrets.SecretDiskMemoryManager.
func (u *NginxConfigBuilder) Build(cfg dataplane.Configuration) (*NginxConfig, error) {
	confDirectory := u.generateConfigDirectory(cfg)
	auxDirectory := u.generateAuxConfigDirectory()

	directories := []*proto.Directory{
		convertToProtoDirectory(confDirectory),
		convertToProtoDirectory(auxDirectory),
	}

	zconfig, err := u.generateZippedFile(confDirectory)
	if err != nil {
		return nil, err
	}

	zaux, err := u.generateZippedFile(auxDirectory)
	if err != nil {
		return nil, err
	}

	return &NginxConfig{
		ID:          fmt.Sprintf("%d", cfg.Generation),
		Config:      zconfig,
		Aux:         zaux,
		Directories: directories,
	}, nil
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

func (u *NginxConfigBuilder) generateConfigDirectory(cfg dataplane.Configuration) directory {
	return directory{
		prefix: confPrefix,
		files: []file{
			{
				path:     nginxConfFilePath,
				mode:     secretsFileMode,
				contents: u.generator.GenerateMainConf(cfg.Generation),
			},
			{
				path:     httpConfFilePath,
				mode:     secretsFileMode,
				contents: u.generator.GenerateHTTPConf(cfg),
			},
		},
	}
}

func (u *NginxConfigBuilder) generateAuxConfigDirectory() directory {
	secretFiles := u.secretMemMgr.GetAllRequestedSecrets()

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

func (u *NginxConfigBuilder) generateZippedFile(dir directory) (*proto.ZippedFile, error) {
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
