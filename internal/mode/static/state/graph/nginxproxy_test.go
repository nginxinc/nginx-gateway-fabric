package graph

import (
	"testing"

	. "github.com/onsi/gomega"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
)

func TestValidateTracing(t *testing.T) {
	gwNsName := "nginx-gateway/nginx-gateway"
	validTracingAllFieldsSet := &ngfAPI.Tracing{
		Enable:     helpers.GetPointer(true),
		Endpoint:   "1.2.3.4:123",
		Interval:   helpers.GetPointer("10s"),
		BatchSize:  helpers.GetPointer(int32(1024)),
		BatchCount: helpers.GetPointer(int32(8)),
	}

	validTracingMinFieldsSet := &ngfAPI.Tracing{
		Endpoint: "1.2.3.4:123",
	}

	invalidTracingNoEndpoint := &ngfAPI.Tracing{
		Enable:   helpers.GetPointer(true),
		Endpoint: "",
		Interval: helpers.GetPointer("187plk"),
	}

	invalidTracingWrongEpFmt := &ngfAPI.Tracing{
		Enable:     helpers.GetPointer(true),
		Endpoint:   "ak,sjufgcebte",
		Interval:   helpers.GetPointer("187plk"),
		BatchSize:  helpers.GetPointer(int32(512)),
		BatchCount: helpers.GetPointer(int32(4)),
	}

	expectedTracingValid := &Tracing{
		Enabled:     true,
		Endpoint:    "1.2.3.4:123",
		Interval:    "10s",
		BatchSize:   1024,
		BatchCount:  8,
		ServiceName: gwNsName + ":ngf",
	}

	expectedTracingValidDefaults := &Tracing{
		Enabled:     false,
		Endpoint:    "1.2.3.4:123",
		Interval:    "5s",
		BatchSize:   512,
		BatchCount:  4,
		ServiceName: gwNsName + ":ngf",
	}

	tests := []struct {
		tracing         *ngfAPI.Tracing
		expectedTracing *Tracing
		name            string
		expectedErrors  bool
	}{
		{
			tracing:         validTracingAllFieldsSet,
			expectedErrors:  false,
			expectedTracing: expectedTracingValid,
			name:            "valid tracing all fields set",
		},
		{
			tracing:         validTracingMinFieldsSet,
			expectedErrors:  false,
			expectedTracing: expectedTracingValidDefaults,
			name:            "valid tracing minimum fields set",
		},
		{
			tracing:         invalidTracingNoEndpoint,
			expectedErrors:  true,
			expectedTracing: nil,
			name:            "invalid tracing no endpoint set",
		},
		{
			tracing:         invalidTracingWrongEpFmt,
			expectedErrors:  true,
			expectedTracing: nil,
			name:            "invalid tracing incorrect endpoint format",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			tracing, valErrs := validateTracing(test.tracing, gwNsName)

			g.Expect(helpers.Diff(test.expectedTracing, tracing)).To(BeEmpty())
			g.Expect(helpers.Diff(test.expectedErrors, len(valErrs) != 0)).To(BeEmpty())
		})
	}
}
