package static

import (
	"context"
	"testing"

	ctlrZap "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestTelemetry(t *testing.T) {
	logger := ctlrZap.New()

	logger.Info("Starting exporter")

	reporter := &TelemetryReporter{
		Logger: logger,
	}

	ctx := context.Background()
	reporter.Start(ctx)
}
