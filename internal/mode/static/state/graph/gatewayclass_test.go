package graph

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	v1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/gatewayclass"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	staticConds "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/conditions"
)

func TestProcessGatewayClasses(t *testing.T) {
	gcName := "test-gc"
	ctlrName := "test-ctlr"
	winner := &v1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: gcName,
		},
		Spec: v1.GatewayClassSpec{
			ControllerName: v1.GatewayController(ctlrName),
		},
	}
	ignored := &v1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-gc-ignored",
		},
		Spec: v1.GatewayClassSpec{
			ControllerName: v1.GatewayController(ctlrName),
		},
	}

	tests := []struct {
		expected processedGatewayClasses
		gcs      map[types.NamespacedName]*v1.GatewayClass
		name     string
		exists   bool
	}{
		{
			gcs:      nil,
			expected: processedGatewayClasses{},
			name:     "no gatewayclasses",
		},
		{
			gcs: map[types.NamespacedName]*v1.GatewayClass{
				{Name: gcName}: winner,
			},
			expected: processedGatewayClasses{
				Winner: winner,
			},
			exists: true,
			name:   "one valid gatewayclass",
		},
		{
			gcs: map[types.NamespacedName]*v1.GatewayClass{
				{Name: gcName}: {
					ObjectMeta: metav1.ObjectMeta{
						Name: gcName,
					},
					Spec: v1.GatewayClassSpec{
						ControllerName: v1.GatewayController("not ours"),
					},
				},
			},
			expected: processedGatewayClasses{},
			exists:   true,
			name:     "one valid gatewayclass, but references wrong controller",
		},
		{
			gcs: map[types.NamespacedName]*v1.GatewayClass{
				{Name: ignored.Name}: ignored,
			},
			expected: processedGatewayClasses{
				Ignored: map[types.NamespacedName]*v1.GatewayClass{
					client.ObjectKeyFromObject(ignored): ignored,
				},
			},
			name: "one non-referenced gatewayclass with our controller",
		},
		{
			gcs: map[types.NamespacedName]*v1.GatewayClass{
				{Name: "completely ignored"}: {
					Spec: v1.GatewayClassSpec{
						ControllerName: v1.GatewayController("not ours"),
					},
				},
			},
			expected: processedGatewayClasses{},
			name:     "one non-referenced gatewayclass without our controller",
		},
		{
			gcs: map[types.NamespacedName]*v1.GatewayClass{
				{Name: gcName}:       winner,
				{Name: ignored.Name}: ignored,
			},
			expected: processedGatewayClasses{
				Winner: winner,
				Ignored: map[types.NamespacedName]*v1.GatewayClass{
					client.ObjectKeyFromObject(ignored): ignored,
				},
			},
			exists: true,
			name:   "one valid gateway class and non-referenced gatewayclass",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)
			result, exists := processGatewayClasses(test.gcs, gcName, ctlrName)
			g.Expect(helpers.Diff(test.expected, result)).To(BeEmpty())
			g.Expect(exists).To(Equal(test.exists))
		})
	}
}

func TestBuildGatewayClass(t *testing.T) {
	validGC := &v1.GatewayClass{}

	invalidGC := &v1.GatewayClass{
		Spec: v1.GatewayClassSpec{
			ParametersRef: &v1.ParametersReference{},
		},
	}

	validCRDs := map[types.NamespacedName]*metav1.PartialObjectMetadata{
		{Name: "gateways.gateway.networking.k8s.io"}: {
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					gatewayclass.BundleVersionAnnotation: gatewayclass.SupportedVersion,
				},
			},
		},
	}

	invalidCRDs := map[types.NamespacedName]*metav1.PartialObjectMetadata{
		{Name: "gateways.gateway.networking.k8s.io"}: {
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					gatewayclass.BundleVersionAnnotation: "v99.0.0",
				},
			},
		},
	}

	tests := []struct {
		gc          *v1.GatewayClass
		crdMetadata map[types.NamespacedName]*metav1.PartialObjectMetadata
		expected    *GatewayClass
		name        string
	}{
		{
			gc:          validGC,
			crdMetadata: validCRDs,
			expected: &GatewayClass{
				Source: validGC,
				Valid:  true,
			},
			name: "valid gatewayclass",
		},
		{
			gc:       nil,
			expected: nil,
			name:     "no gatewayclass",
		},
		{
			gc:          invalidGC,
			crdMetadata: validCRDs,
			expected: &GatewayClass{
				Source: invalidGC,
				Valid:  false,
				Conditions: []conditions.Condition{
					staticConds.NewGatewayClassInvalidParameters(
						"spec.parametersRef: Forbidden: parametersRef is not supported",
					),
				},
			},
			name: "invalid gatewayclass; parameters ref",
		},
		{
			gc:          validGC,
			crdMetadata: invalidCRDs,
			expected: &GatewayClass{
				Source:     validGC,
				Valid:      false,
				Conditions: conditions.NewGatewayClassUnsupportedVersion(gatewayclass.SupportedVersion),
			},
			name: "invalid gatewayclass; unsupported version",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			result := buildGatewayClass(test.gc, test.crdMetadata)
			g.Expect(helpers.Diff(test.expected, result)).To(BeEmpty())
		})
	}
}
