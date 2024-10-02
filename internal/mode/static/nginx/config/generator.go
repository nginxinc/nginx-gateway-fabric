package config

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ngfConfig "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/config"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config/policies"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config/policies/clientsettings"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config/policies/observability"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/file"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/telemetry"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . Generator

const (
	// configFolder is the folder where NGINX configuration files are stored.
	configFolder = "/etc/nginx"

	// httpFolder is the folder where NGINX HTTP configuration files are stored.
	httpFolder = configFolder + "/conf.d"

	// streamFolder is the folder where NGINX Stream configuration files are stored.
	streamFolder = configFolder + "/stream-conf.d"

	// mainIncludesFolder is the folder where NGINX main context configuration files are stored.
	// For example, these files include load_module directives and snippets that target the main context.
	mainIncludesFolder = configFolder + "/main-includes"

	// secretsFolder is the folder where secrets (like TLS certs/keys) are stored.
	secretsFolder = configFolder + "/secrets"

	// includesFolder is the folder where are all include files are stored.
	includesFolder = configFolder + "/includes"

	// httpConfigFile is the path to the configuration file with HTTP configuration.
	httpConfigFile = httpFolder + "/http.conf"

	// streamConfigFile is the path to the configuration file with Stream configuration.
	streamConfigFile = streamFolder + "/stream.conf"

	// configVersionFile is the path to the config version configuration file.
	configVersionFile = httpFolder + "/config-version.conf"

	// httpMatchVarsFile is the path to the http_match pairs configuration file.
	httpMatchVarsFile = httpFolder + "/matches.json"

	// mainIncludesConfigFile is the path to the file containing NGINX configuration in the main context.
	mainIncludesConfigFile = mainIncludesFolder + "/main.conf"

	// mgmtIncludesFile is the path to the file containing the NGINX Plus mgmt config.
	mgmtIncludesFile = mainIncludesFolder + "/mgmt.conf"
)

// ConfigFolders is a list of folders where NGINX configuration files are stored.
// Volumes here also need to be added to our crossplane ephemeral test container.
var ConfigFolders = []string{httpFolder, secretsFolder, includesFolder, mainIncludesFolder, streamFolder}

// Generator generates NGINX configuration files.
// This interface is used for testing purposes only.
type Generator interface {
	// Generate generates NGINX configuration files from internal representation.
	Generate(configuration dataplane.Configuration) []file.File
}

// GeneratorImpl is an implementation of Generator.
//
// It generates files to be written to the ConfigFolders locations, which must exist and available for writing.
//
// It also expects that the main NGINX configuration file nginx.conf is located in configFolder and nginx.conf
// includes (https://nginx.org/en/docs/ngx_core_module.html#include) the files from other folders.
type GeneratorImpl struct {
	k8sClientReader   client.Reader
	usageReportConfig *ngfConfig.UsageReportConfig
	podConfig         *ngfConfig.GatewayPodConfig
	logger            logr.Logger
	plus              bool
}

// NewGeneratorImpl creates a new GeneratorImpl.
func NewGeneratorImpl(
	plus bool,
	usageReportConfig *ngfConfig.UsageReportConfig,
	podConfig *ngfConfig.GatewayPodConfig,
	reader client.Reader,
	logger logr.Logger,
) GeneratorImpl {
	return GeneratorImpl{
		plus:              plus,
		usageReportConfig: usageReportConfig,
		podConfig:         podConfig,
		k8sClientReader:   reader,
		logger:            logger,
	}
}

type executeResult struct {
	dest string
	data []byte
}

// executeFunc is a function that generates NGINX configuration from internal representation.
type executeFunc func(configuration dataplane.Configuration) []executeResult

// Generate generates NGINX configuration files from internal representation.
// It is the responsibility of the caller to validate the configuration before calling this function.
// In case of invalid configuration, NGINX will fail to reload or could be configured with malicious configuration.
// To validate, use the validators from the validation package.
func (g GeneratorImpl) Generate(conf dataplane.Configuration) []file.File {
	files := make([]file.File, 0)

	for id, pair := range conf.SSLKeyPairs {
		files = append(files, generatePEM(id, pair.Cert, pair.Key))
	}

	policyGenerator := policies.NewCompositeGenerator(
		clientsettings.NewGenerator(),
		observability.NewGenerator(conf.Telemetry),
	)

	files = append(files, g.executeConfigTemplates(conf, policyGenerator)...)

	for id, bundle := range conf.CertBundles {
		files = append(files, generateCertBundle(id, bundle))
	}

	return files
}

func (g GeneratorImpl) executeConfigTemplates(
	conf dataplane.Configuration,
	generator policies.Generator,
) []file.File {
	fileBytes := make(map[string][]byte)

	for _, execute := range g.getExecuteFuncs(generator) {
		results := execute(conf)
		for _, res := range results {
			fileBytes[res.dest] = append(fileBytes[res.dest], res.data...)
		}
	}

	var plusFileCount int
	if g.plus {
		plusFileCount = 2 // mgmt.conf, deployment_ctx.json
	}

	files := make([]file.File, 0, len(fileBytes)+plusFileCount)
	for fp, bytes := range fileBytes {
		files = append(files, file.File{
			Path:    fp,
			Content: bytes,
			Type:    file.TypeRegular,
		})
	}

	if g.plus {
		files = append(files, g.generateMgmtFiles()...)
	}

	return files
}

func (g GeneratorImpl) getExecuteFuncs(generator policies.Generator) []executeFunc {
	return []executeFunc{
		executeMainConfig,
		executeBaseHTTPConfig,
		g.newExecuteServersFunc(generator),
		g.executeUpstreams,
		executeSplitClients,
		executeMaps,
		executeTelemetry,
		g.executeStreamServers,
		g.executeStreamUpstreams,
		executeStreamMaps,
		executeVersion,
	}
}

func generatePEM(id dataplane.SSLKeyPairID, cert []byte, key []byte) file.File {
	c := make([]byte, 0, len(cert)+len(key)+1)
	c = append(c, cert...)
	c = append(c, '\n')
	c = append(c, key...)

	return file.File{
		Content: c,
		Path:    generatePEMFileName(id),
		Type:    file.TypeSecret,
	}
}

func generatePEMFileName(id dataplane.SSLKeyPairID) string {
	return filepath.Join(secretsFolder, string(id)+".pem")
}

func generateCertBundle(id dataplane.CertBundleID, cert []byte) file.File {
	return file.File{
		Content: cert,
		Path:    generateCertBundleFileName(id),
		Type:    file.TypeRegular,
	}
}

func generateCertBundleFileName(id dataplane.CertBundleID) string {
	return filepath.Join(secretsFolder, string(id)+".crt")
}

type mgmtConf struct {
	Endpoint            string
	Resolver            string
	DeploymentCtxFile   string
	SkipVerify          bool
	CACertExists        bool
	ClientSSLCertExists bool
}

// generateMgmtFile generates the NGINX Plus configuration file for the mgmt block. As part of this,
// it writes the deployment context file that is referenced in the mgmt block.
func (g GeneratorImpl) generateMgmtFiles() []file.File {
	var files []file.File

	podNSName := types.NamespacedName{
		Namespace: g.podConfig.Namespace,
		Name:      g.podConfig.Name,
	}

	conf := mgmtConf{
		Endpoint:            g.usageReportConfig.Endpoint,
		Resolver:            g.usageReportConfig.Resolver,
		SkipVerify:          g.usageReportConfig.SkipVerify,
		CACertExists:        g.usageReportConfig.CASecretName != "",
		ClientSSLCertExists: g.usageReportConfig.ClientSSLSecretName != "",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	data, err := getDeploymentCtx(ctx, g.k8sClientReader, podNSName)
	if err != nil {
		g.logger.Error(err, "error building deployment context for mgmt block")
	} else {
		deploymentCtxFile := file.File{
			Content: data,
			Path:    mainIncludesFolder + "/deployment_ctx.json",
			Type:    file.TypeRegular,
		}

		conf.DeploymentCtxFile = deploymentCtxFile.Path
		files = []file.File{deploymentCtxFile}
	}

	mgmtBlockFile := file.File{
		Content: executeMgmtIncludesConfig(conf),
		Path:    mgmtIncludesFile,
		Type:    file.TypeRegular,
	}

	return append(files, mgmtBlockFile)
}

type deploymentContext struct {
	Integration      string `json:"integration"`
	ClusterID        string `json:"cluster_id"`
	InstallationID   string `json:"installation_id"`
	ClusterNodeCount int    `json:"cluster_node_count"`
}

func getDeploymentCtx(
	ctx context.Context,
	reader client.Reader,
	podNSName types.NamespacedName,
) ([]byte, error) {
	clusterInfo, err := telemetry.CollectClusterInformation(ctx, reader)
	if err != nil {
		return nil, fmt.Errorf("error getting cluster information")
	}

	replicaSet, err := telemetry.GetPodReplicaSet(ctx, reader, podNSName)
	if err != nil {
		return nil, fmt.Errorf("failed to get replica set for pod %v: %w", podNSName, err)
	}

	deploymentID, err := telemetry.GetDeploymentID(replicaSet)
	if err != nil {
		return nil, fmt.Errorf("failed to get NGF deploymentID: %w", err)
	}

	depCtx := &deploymentContext{
		Integration:      "ngf",
		ClusterID:        clusterInfo.ClusterID,
		ClusterNodeCount: clusterInfo.NodeCount,
		InstallationID:   deploymentID,
	}

	b, err := json.Marshal(depCtx)
	if err != nil {
		return nil, fmt.Errorf("error marshaling json data: %w", err)
	}

	return b, nil
}
