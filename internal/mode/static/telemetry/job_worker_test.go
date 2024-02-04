package telemetry_test

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/telemetry"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/telemetry/telemetryfakes"
)

func TestCreateTelemetryJobWorker(t *testing.T) {
	g := NewWithT(t)

	exporter := &telemetryfakes.FakeExporter{}
	dataCollector := &telemetryfakes.FakeDataCollector{}

	worker := telemetry.CreateTelemetryJobWorker(zap.New(), exporter, dataCollector)

	expData := telemetry.Data{
		ProjectMetadata: telemetry.ProjectMetadata{Name: "NGF", Version: "1.1"},
		NodeCount:       3,
		NGFResourceCounts: telemetry.NGFResourceCounts{
			Gateways:       1,
			GatewayClasses: 1,
			HTTPRoutes:     1,
			Secrets:        1,
			Services:       1,
			Endpoints:      1,
		},
	}
	dataCollector.CollectReturns(expData, nil)

	timeout := 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	worker(ctx)
	_, data := exporter.ExportArgsForCall(0)
	g.Expect(data).To(Equal(expData))
	cancel()
}
