package static

import (
	"testing"

	. "github.com/onsi/gomega"
	apiv1 "k8s.io/api/core/v1"
	discoveryV1 "k8s.io/api/discovery/v1"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gatewayv1alpha3 "sigs.k8s.io/gateway-api/apis/v1alpha3"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/config"
)

func TestPrepareFirstEventBatchPreparerArgs(t *testing.T) {
	t.Parallel()
	const gcName = "nginx"

	partialObjectMetadataList := &metav1.PartialObjectMetadataList{}
	partialObjectMetadataList.SetGroupVersionKind(
		schema.GroupVersionKind{
			Group:   apiext.GroupName,
			Version: "v1",
			Kind:    "CustomResourceDefinition",
		},
	)

	tests := []struct {
		expectedObjects     []client.Object
		expectedObjectLists []client.ObjectList
		name                string
		cfg                 config.Config
	}{
		{
			name: "gwNsName is nil",
			cfg: config.Config{
				GatewayClassName:     gcName,
				GatewayNsName:        nil,
				ExperimentalFeatures: false,
				SnippetsFilters:      false,
			},
			expectedObjects: []client.Object{
				&gatewayv1.GatewayClass{ObjectMeta: metav1.ObjectMeta{Name: "nginx"}},
			},
			expectedObjectLists: []client.ObjectList{
				&apiv1.ServiceList{},
				&apiv1.SecretList{},
				&apiv1.NamespaceList{},
				&discoveryV1.EndpointSliceList{},
				&gatewayv1.HTTPRouteList{},
				&gatewayv1.GatewayList{},
				&gatewayv1beta1.ReferenceGrantList{},
				&ngfAPI.NginxProxyList{},
				&gatewayv1.GRPCRouteList{},
				partialObjectMetadataList,
				&ngfAPI.ClientSettingsPolicyList{},
				&ngfAPI.ObservabilityPolicyList{},
			},
		},
		{
			name: "gwNsName is not nil",
			cfg: config.Config{
				GatewayClassName: gcName,
				GatewayNsName: &types.NamespacedName{
					Namespace: "test",
					Name:      "my-gateway",
				},
				ExperimentalFeatures: false,
				SnippetsFilters:      false,
			},
			expectedObjects: []client.Object{
				&gatewayv1.GatewayClass{ObjectMeta: metav1.ObjectMeta{Name: "nginx"}},
				&gatewayv1.Gateway{ObjectMeta: metav1.ObjectMeta{Name: "my-gateway", Namespace: "test"}},
			},
			expectedObjectLists: []client.ObjectList{
				&apiv1.ServiceList{},
				&apiv1.SecretList{},
				&apiv1.NamespaceList{},
				&discoveryV1.EndpointSliceList{},
				&gatewayv1.HTTPRouteList{},
				&gatewayv1beta1.ReferenceGrantList{},
				&ngfAPI.NginxProxyList{},
				&gatewayv1.GRPCRouteList{},
				partialObjectMetadataList,
				&ngfAPI.ClientSettingsPolicyList{},
				&ngfAPI.ObservabilityPolicyList{},
			},
		},
		{
			name: "gwNsName is not nil and experimental enabled",
			cfg: config.Config{
				GatewayClassName: gcName,
				GatewayNsName: &types.NamespacedName{
					Namespace: "test",
					Name:      "my-gateway",
				},
				ExperimentalFeatures: true,
				SnippetsFilters:      false,
			},
			expectedObjects: []client.Object{
				&gatewayv1.GatewayClass{ObjectMeta: metav1.ObjectMeta{Name: "nginx"}},
				&gatewayv1.Gateway{ObjectMeta: metav1.ObjectMeta{Name: "my-gateway", Namespace: "test"}},
			},
			expectedObjectLists: []client.ObjectList{
				&apiv1.ServiceList{},
				&apiv1.SecretList{},
				&apiv1.NamespaceList{},
				&apiv1.ConfigMapList{},
				&discoveryV1.EndpointSliceList{},
				&gatewayv1.HTTPRouteList{},
				&gatewayv1beta1.ReferenceGrantList{},
				&ngfAPI.NginxProxyList{},
				partialObjectMetadataList,
				&gatewayv1alpha3.BackendTLSPolicyList{},
				&gatewayv1alpha2.TLSRouteList{},
				&gatewayv1.GRPCRouteList{},
				&ngfAPI.ClientSettingsPolicyList{},
				&ngfAPI.ObservabilityPolicyList{},
			},
		},
		{
			name: "gwNsName is not nil and snippets filters enabled",
			cfg: config.Config{
				GatewayClassName: gcName,
				GatewayNsName: &types.NamespacedName{
					Namespace: "test",
					Name:      "my-gateway",
				},
				ExperimentalFeatures: false,
				SnippetsFilters:      true,
			},
			expectedObjects: []client.Object{
				&gatewayv1.GatewayClass{ObjectMeta: metav1.ObjectMeta{Name: "nginx"}},
				&gatewayv1.Gateway{ObjectMeta: metav1.ObjectMeta{Name: "my-gateway", Namespace: "test"}},
			},
			expectedObjectLists: []client.ObjectList{
				&apiv1.ServiceList{},
				&apiv1.SecretList{},
				&apiv1.NamespaceList{},
				&discoveryV1.EndpointSliceList{},
				&gatewayv1.HTTPRouteList{},
				&gatewayv1beta1.ReferenceGrantList{},
				&ngfAPI.NginxProxyList{},
				partialObjectMetadataList,
				&gatewayv1.GRPCRouteList{},
				&ngfAPI.ClientSettingsPolicyList{},
				&ngfAPI.ObservabilityPolicyList{},
				&ngfAPI.SnippetsFilterList{},
			},
		},
		{
			name: "gwNsName is not nil, experimental and snippets filters enabled",
			cfg: config.Config{
				GatewayClassName: gcName,
				GatewayNsName: &types.NamespacedName{
					Namespace: "test",
					Name:      "my-gateway",
				},
				ExperimentalFeatures: true,
				SnippetsFilters:      true,
			},
			expectedObjects: []client.Object{
				&gatewayv1.GatewayClass{ObjectMeta: metav1.ObjectMeta{Name: "nginx"}},
				&gatewayv1.Gateway{ObjectMeta: metav1.ObjectMeta{Name: "my-gateway", Namespace: "test"}},
			},
			expectedObjectLists: []client.ObjectList{
				&apiv1.ServiceList{},
				&apiv1.SecretList{},
				&apiv1.NamespaceList{},
				&apiv1.ConfigMapList{},
				&discoveryV1.EndpointSliceList{},
				&gatewayv1.HTTPRouteList{},
				&gatewayv1beta1.ReferenceGrantList{},
				&ngfAPI.NginxProxyList{},
				partialObjectMetadataList,
				&gatewayv1alpha3.BackendTLSPolicyList{},
				&gatewayv1alpha2.TLSRouteList{},
				&gatewayv1.GRPCRouteList{},
				&ngfAPI.ClientSettingsPolicyList{},
				&ngfAPI.ObservabilityPolicyList{},
				&ngfAPI.SnippetsFilterList{},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			objects, objectLists := prepareFirstEventBatchPreparerArgs(test.cfg)

			g.Expect(objects).To(ConsistOf(test.expectedObjects))
			g.Expect(objectLists).To(ConsistOf(test.expectedObjectLists))
		})
	}
}

func TestGetMetricsOptions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		expectedOptions metricsserver.Options
		metricsConfig   config.MetricsConfig
	}{
		{
			name:            "Metrics disabled",
			metricsConfig:   config.MetricsConfig{Enabled: false},
			expectedOptions: metricsserver.Options{BindAddress: "0"},
		},
		{
			name: "Metrics enabled, not secure",
			metricsConfig: config.MetricsConfig{
				Port:    9113,
				Enabled: true,
				Secure:  false,
			},
			expectedOptions: metricsserver.Options{
				SecureServing: false,
				BindAddress:   ":9113",
			},
		},
		{
			name: "Metrics enabled, secure",
			metricsConfig: config.MetricsConfig{
				Port:    9113,
				Enabled: true,
				Secure:  true,
			},
			expectedOptions: metricsserver.Options{
				SecureServing: true,
				BindAddress:   ":9113",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			metricsServerOptions := getMetricsOptions(test.metricsConfig)

			g.Expect(metricsServerOptions).To(Equal(test.expectedOptions))
		})
	}
}
