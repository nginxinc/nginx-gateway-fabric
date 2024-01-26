package usage_test

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/config"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/usage"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/usage/usagefakes"
)

func TestCreateUsageJobWorker(t *testing.T) {
	g := NewWithT(t)

	replicas := int32(1)
	ngfReplicaSet := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "nginx-gateway",
			Name:      "ngf-replicaset",
		},
		Spec: appsv1.ReplicaSetSpec{
			Replicas: &replicas,
		},
	}

	ngfPod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "nginx-gateway",
			Name:      "ngf-pod",
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind: "ReplicaSet",
					Name: "ngf-replicaset",
				},
			},
		},
	}

	kubeSystem := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: metav1.NamespaceSystem,
			UID:  "1234abcd",
		},
	}

	k8sClient := fake.NewFakeClient(&v1.Node{}, ngfReplicaSet, ngfPod, kubeSystem)
	reporter := &usagefakes.FakeReporter{}

	worker := usage.CreateUsageJobWorker(
		zap.New(),
		k8sClient,
		reporter,
		config.Config{
			GatewayPodConfig: config.GatewayPodConfig{
				Namespace: "nginx-gateway",
				Name:      "ngf-pod",
			},
			UsageReportConfig: &config.UsageReportConfig{
				ClusterDisplayName: "my-cluster",
			},
		},
	)

	expData := usage.ClusterDetails{
		Metadata: usage.Metadata{
			UID:         "1234abcd",
			DisplayName: "my-cluster",
		},
		NodeCount: 1,
		PodDetails: usage.PodDetails{
			CurrentPodCounts: usage.CurrentPodsCount{
				PodCount: 1,
			},
		},
	}

	timeout := 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	worker(ctx)
	_, data := reporter.ReportArgsForCall(0)
	g.Expect(data).To(Equal(expData))
}
