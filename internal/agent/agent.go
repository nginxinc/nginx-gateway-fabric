package agent

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-kit/log"
	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/exporter-toolkit/web"
	"google.golang.org/grpc"
	ctlr "sigs.k8s.io/controller-runtime"

	"github.com/nginxinc/nginx-gateway-fabric/internal/agent/file"
	"github.com/nginxinc/nginx-gateway-fabric/internal/agent/runtime"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	"github.com/nginxinc/nginx-gateway-fabric/internal/grpc/controlplane"
	"github.com/nginxinc/nginx-gateway-fabric/internal/grpc/sdk/client"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/metrics/collectors"
	ngxcfg "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config"
)

type Config struct {
	Logger               logr.Logger
	ControlPlaneEndpoint string
}

func Start(cfg Config) error {
	ctx := ctlr.SetupSignalHandler()

	cfg.Logger.Info("starting agent",
		"controlPlaneEndpoint", cfg.ControlPlaneEndpoint,
	)

	nginxFileMgr := file.NewManagerImpl(
		cfg.Logger.WithName("nginxFileManager"),
		file.NewStdLibOSFileManager(),
	)

	// Clear the configuration folders to ensure that no files are left over in case the control plane was restarted
	// (this assumes the folders are in a shared volume).
	removedPaths, err := file.ClearFolders(file.NewStdLibOSFileManager(), ngxcfg.ConfigFolders)
	for _, path := range removedPaths {
		cfg.Logger.Info("removed configuration file", "path", path)
	}
	if err != nil {
		return fmt.Errorf("cannot clear NGINX configuration folders: %w", err)
	}

	// Ensure NGINX is running before registering metrics & starting the manager.
	if err := runtime.EnsureNginxRunning(ctx); err != nil {
		return fmt.Errorf("NGINX is not running: %w", err)
	}

	constLabels := map[string]string{"class": "nginx"}
	ngxruntimeCollector := collectors.NewManagerMetricsCollector(constLabels)

	ngxCollector, err := collectors.NewNginxMetricsCollector(constLabels)
	if err != nil {
		return err
	}

	prometheus.MustRegister(ngxruntimeCollector, ngxCollector)
	http.Handle("/metrics", promhttp.Handler())

	nginxMgr := runtime.NewManagerImpl(ngxruntimeCollector)

	agent := newAgent(
		cfg.Logger.WithName("agent"),
		cfg.ControlPlaneEndpoint,
		nginxFileMgr,
		nginxMgr,
	)

	go func() {
		srv := &http.Server{
			ReadHeaderTimeout: 5 * time.Second,
		}

		promLogger, err := newPromLogger()
		if err != nil {
			cfg.Logger.Error(err, "failed to create Prometheus logger")
			return
		}

		flagConfig := web.FlagConfig{
			WebListenAddresses: helpers.GetPointer([]string{":9113"}),
			WebConfigFile:      helpers.GetPointer(""),
		}
		if err := web.ListenAndServe(srv, &flagConfig, promLogger); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				cfg.Logger.Info("Prometheus exporter stopped")
				return
			}
			cfg.Logger.Error(err, "Prometheus exporter failed")
		}
		return
	}()

	return agent.start(ctx)
}

// newPromLogger creates a Prometheus logger that implements to go-kit log.Logger that the prometheus exporter requires.
func newPromLogger() (log.Logger, error) {
	logFormat := &promlog.AllowedFormat{}

	if err := logFormat.Set("json"); err != nil {
		return nil, err
	}

	logConfig := &promlog.Config{Format: logFormat}
	return promlog.New(logConfig), nil
}

type agent struct {
	logger       logr.Logger
	endpoint     string
	nginxFileMgr file.Manager
	nginxMgr     runtime.Manager
}

func newAgent(
	logger logr.Logger,
	endpoint string,
	nginxFileMgr file.Manager,
	nginxMgr runtime.Manager,
) *agent {
	return &agent{
		logger:       logger,
		endpoint:     endpoint,
		nginxFileMgr: nginxFileMgr,
		nginxMgr:     nginxMgr,
	}
}

func (a *agent) start(agentCtx context.Context) error {
	ctx, cancel := context.WithCancel(agentCtx)
	defer cancel()

	// sleep for 10 seconds to allow the control plane to start
	// don't judge the POC code
	time.Sleep(10 * time.Second)

	conn, err := grpc.Dial(a.endpoint, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	grpcClient := controlplane.NewControlPlaneClient(conn)

	applyResultCh := make(chan client.ApplyResult)
	defer close(applyResultCh)

	c := client.NewClient(grpcClient, applyResultCh)

	done := make(chan struct{})

	go func(ctx context.Context) {
		defer close(done)

		for {
			a.logger.Info("waiting for config")

			select {
			case <-ctx.Done():
				return
			case cfg := <-c.ConfigCh():
				a.logger.Info("received config", "generation", cfg.Generation)

				err := a.nginxFileMgr.ReplaceFiles(cfg.Files)
				if err != nil {
					a.logger.Error(err, "failed to replace files")

					select {
					case applyResultCh <- client.ApplyResult{
						Generation: cfg.Generation,
						Success:    false,
					}:
					case <-ctx.Done():
						return
					}
					continue
				}

				a.logger.Info("files replaced")

				err = a.nginxMgr.Reload(ctx, int(cfg.Generation))
				if err != nil {
					a.logger.Error(err, "failed to reload nginx")

					select {
					case applyResultCh <- client.ApplyResult{
						Generation: cfg.Generation,
						Success:    false,
					}:
					case <-ctx.Done():
						return
					}

					continue
				}

				a.logger.Info("nginx reloaded")

				select {
				case applyResultCh <- client.ApplyResult{
					Generation: cfg.Generation,
					Success:    true,
				}:
				case <-ctx.Done():
					return
				}
			}
		}
	}(ctx)

	returnErr := c.Start(ctx)
	if returnErr != nil {
		a.logger.Error(returnErr, "failed to start c")
	}
	cancel()
	<-done
	return returnErr
}
