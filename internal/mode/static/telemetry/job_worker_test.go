package telemetry_test

import (
	"context"
	"testing"
	"time"

	tel "github.com/nginxinc/telemetry-exporter/pkg/telemetry"
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
		Data: tel.Data{
			ProjectName: "NGF",
		},
	}
	dataCollector.CollectReturns(expData, nil)

	timeout := 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	worker(ctx)
	_, data := exporter.ExportArgsForCall(0)
	g.Expect(data).To(Equal(&expData))
}
