package main

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	ctlr "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/licensing"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/file"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
)

type copyFiles struct {
	destDirName  string
	srcFileNames []string
}

type initializeConfig struct {
	controllerPodNSName types.NamespacedName
	fileManager         file.OSFileManager
	logger              logr.Logger
	copy                copyFiles
	plus                bool
}

func initialize(cfg initializeConfig) error {
	for _, src := range cfg.copy.srcFileNames {
		if err := copyFile(cfg.fileManager, src, cfg.copy.destDirName); err != nil {
			return err
		}
	}

	if !cfg.plus {
		cfg.logger.Info("Finished initializing configuration")
		return nil
	}

	clusterCfg := ctlr.GetConfigOrDie()
	k8sReader, err := client.New(clusterCfg, client.Options{})
	if err != nil {
		return fmt.Errorf("unable to initialize k8s client: %w", err)
	}

	dcc := licensing.NewDeploymentContextCollector(licensing.DeploymentContextCollectorConfig{
		K8sClientReader: k8sReader,
		PodNSName:       cfg.controllerPodNSName,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	depCtx, err := dcc.Collect(ctx, cfg.logger.WithName("deployCtxCollector"))
	if err != nil {
		return fmt.Errorf("failed to collect deployment context: %w", err)
	}

	cfg.logger.Info("Deployment context collected", "deployment context", depCtx)

	if err := writeDeploymentContextFile(cfg.fileManager, depCtx); err != nil {
		return fmt.Errorf("failed to write deployment context file: %w", err)
	}

	cfg.logger.Info("Finished initializing configuration")

	return nil
}

func writeDeploymentContextFile(osFileManager file.OSFileManager, depCtx dataplane.DeploymentContext) error {
	depCtxFile, err := config.GenerateDeploymentContextFile(depCtx)
	if err != nil {
		return fmt.Errorf("failed to generate deployment context file: %w", err)
	}

	if err := file.WriteFile(osFileManager, depCtxFile); err != nil {
		return fmt.Errorf("failed to write deployment context file: %w", err)
	}

	return nil
}

func copyFile(osFileManager file.OSFileManager, src, dest string) error {
	srcFile, err := osFileManager.Open(src)
	if err != nil {
		return fmt.Errorf("error opening source file: %w", err)
	}
	defer srcFile.Close()

	destFile, err := osFileManager.Create(filepath.Join(dest, filepath.Base(src)))
	if err != nil {
		return fmt.Errorf("error creating destination file: %w", err)
	}
	defer destFile.Close()

	if err := osFileManager.Copy(destFile, srcFile); err != nil {
		return fmt.Errorf("error copying file contents: %w", err)
	}

	return nil
}
