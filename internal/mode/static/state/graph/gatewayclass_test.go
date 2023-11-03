package graph

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	staticConds "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/conditions"
)

func TestProcessGatewayClasses(t *testing.T) {
	gcName := "test-gc"
	ctlrName := "test-ctlr"
	winner := &v1beta1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: gcName,
		},
		Spec: v1beta1.GatewayClassSpec{
			ControllerName: v1beta1.GatewayController(ctlrName),
		},
	}
	ignored := &v1beta1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-gc-ignored",
		},
		Spec: v1beta1.GatewayClassSpec{
			ControllerName: v1beta1.GatewayController(ctlrName),
		},
	}

	tests := []struct {
		expected processedGatewayClasses
		gcs      map[types.NamespacedName]*v1beta1.GatewayClass
		name     string
		exists   bool
	}{
		{
			gcs:      nil,
			expected: processedGatewayClasses{},
			name:     "no gatewayclasses",
		},
		{
			gcs: map[types.NamespacedName]*v1beta1.GatewayClass{
				{Name: gcName}: winner,
			},
			expected: processedGatewayClasses{
				Winner: winner,
			},
			exists: true,
			name:   "one valid gatewayclass",
		},
		{
			gcs: map[types.NamespacedName]*v1beta1.GatewayClass{
				{Name: gcName}: {
					ObjectMeta: metav1.ObjectMeta{
						Name: gcName,
					},
					Spec: v1beta1.GatewayClassSpec{
						ControllerName: v1beta1.GatewayController("not ours"),
					},
				},
			},
			expected: processedGatewayClasses{},
			exists:   true,
			name:     "one valid gatewayclass, but references wrong controller",
		},
		{
			gcs: map[types.NamespacedName]*v1beta1.GatewayClass{
				{Name: ignored.Name}: ignored,
			},
			expected: processedGatewayClasses{
				Ignored: map[types.NamespacedName]*v1beta1.GatewayClass{
					client.ObjectKeyFromObject(ignored): ignored,
				},
			},
			name: "one non-referenced gatewayclass with our controller",
		},
		{
			gcs: map[types.NamespacedName]*v1beta1.GatewayClass{
				{Name: "completely ignored"}: {
					Spec: v1beta1.GatewayClassSpec{
						ControllerName: v1beta1.GatewayController("not ours"),
					},
				},
			},
			expected: processedGatewayClasses{},
			name:     "one non-referenced gatewayclass without our controller",
		},
		{
			gcs: map[types.NamespacedName]*v1beta1.GatewayClass{
				{Name: gcName}:       winner,
				{Name: ignored.Name}: ignored,
			},
			expected: processedGatewayClasses{
				Winner: winner,
				Ignored: map[types.NamespacedName]*v1beta1.GatewayClass{
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
	validGC := &v1beta1.GatewayClass{}

	gcWithParams := &v1beta1.GatewayClass{
		Spec: v1beta1.GatewayClassSpec{
			ParametersRef: &v1beta1.ParametersReference{
				Kind:      v1beta1.Kind("NginxProxy"),
				Namespace: helpers.GetPointer(v1beta1.Namespace("test")),
				Name:      "does-not-exist",
			},
		},
	}

	gcWithInvalidKind := &v1beta1.GatewayClass{
		Spec: v1beta1.GatewayClassSpec{
			ParametersRef: &v1beta1.ParametersReference{
				Kind:      v1beta1.Kind("Invalid"),
				Namespace: helpers.GetPointer(v1beta1.Namespace("test")),
			},
		},
	}

	gcWithNoNamespace := &v1beta1.GatewayClass{
		Spec: v1beta1.GatewayClassSpec{
			ParametersRef: &v1beta1.ParametersReference{
				Kind: v1beta1.Kind("NginxProxy"),
			},
		},
	}

	tests := []struct {
		gc       *v1beta1.GatewayClass
		np       *ngfAPI.NginxProxy
		expected *GatewayClass
		name     string
	}{
		{
			gc: validGC,
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
			gc: gcWithParams,
			np: &ngfAPI.NginxProxy{
				TypeMeta: metav1.TypeMeta{
					Kind: "NginxProxy",
				},
			},
			expected: &GatewayClass{
				Source: gcWithParams,
				Valid:  true,
			},
			name: "valid gatewayclass with paramsRef",
		},
		{
			gc: gcWithInvalidKind,
			np: &ngfAPI.NginxProxy{
				TypeMeta: metav1.TypeMeta{
					Kind: "NginxProxy",
				},
			},
			expected: &GatewayClass{
				Source: gcWithInvalidKind,
				Valid:  false,
				Conditions: []conditions.Condition{
					staticConds.NewGatewayClassInvalidParameters(
						"spec.parametersRef.kind: Unsupported value: \"Invalid\": supported values: \"NginxProxy\"",
					),
				},
			},
			name: "invalid gatewayclass with unsupported paramsRef Kind",
		},
		{
			gc: gcWithParams,
			expected: &GatewayClass{
				Source: gcWithParams,
				Valid:  false,
				Conditions: []conditions.Condition{
					staticConds.NewGatewayClassInvalidParameters(
						"spec.parametersRef.name: Not found: \"does-not-exist\"",
					),
				},
			},
			name: "invalid gatewayclass with paramsRef resource that doesn't exist",
		},
		{
			gc: gcWithNoNamespace,
			np: &ngfAPI.NginxProxy{
				TypeMeta: metav1.TypeMeta{
					Kind: "NginxProxy",
				},
			},
			expected: &GatewayClass{
				Source: gcWithNoNamespace,
				Valid:  false,
				Conditions: []conditions.Condition{
					staticConds.NewGatewayClassInvalidParameters(
						"spec.parametersRef.namespace: Required value: parametersRef.namespace must be specified for NginxProxy",
					),
				},
			},
			name: "invalid gatewayclass without required paramsRef Namespace",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			result := buildGatewayClass(test.gc, test.np)
			g.Expect(helpers.Diff(test.expected, result)).To(BeEmpty())
		})
	}
}
