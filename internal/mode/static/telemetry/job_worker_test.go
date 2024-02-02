package telemetry_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/telemetry"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/telemetry/telemetryfakes"
)

var _ = Describe("Job Worker", func() {
	var (
		exporter        *telemetryfakes.FakeExporter
		dataCollector   *telemetryfakes.FakeDataCollector
		healthCollector *telemetryfakes.FakeHealthChecker
		expData         telemetry.Data
		worker          func(context.Context)
		readyChannel    chan struct{}
	)
	const timeout = 10 * time.Second

	BeforeEach(func() {
		exporter = &telemetryfakes.FakeExporter{}
		dataCollector = &telemetryfakes.FakeDataCollector{}
		healthCollector = &telemetryfakes.FakeHealthChecker{}

		readyChannel = make(chan struct{})
		healthCollector.GetReadyIfClosedChannelReturns(readyChannel)

		worker = telemetry.CreateTelemetryJobWorker(zap.New(), exporter, dataCollector, healthCollector)

		expData = telemetry.Data{
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
	})

	DescribeTable(
		"Job worker runs without any errors",
		func(sleep time.Duration) {
			ctx, cancel := context.WithTimeout(context.Background(), timeout)

			go func() {
				time.Sleep(sleep)
				close(readyChannel)
			}()

			worker(ctx)
			_, data := exporter.ExportArgsForCall(0)
			Expect(data).To(Equal(expData))
			cancel()
		},
		Entry("Job worker runs with NGF Pod ready", time.Duration(0)),
		Entry("Job worker runs with extended time waiting for NGF Pod to be ready", 1*time.Second),
	)

	When("job worker context is canceled", func() {
		It("should gracefully exist", func() {
			ctx, cancel := context.WithTimeout(context.Background(), timeout)

			go func() {
				worker(ctx)
			}()

			sleep := 500 * time.Millisecond
			time.Sleep(sleep)
			cancel()

			Expect(dataCollector.CollectCallCount()).To(BeZero())
			Expect(exporter.ExportCallCount()).To(BeZero())
		})
	})
})
