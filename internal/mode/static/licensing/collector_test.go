package licensing_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/licensing"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
)

var _ = Describe("DeploymentContextCollector", func() {
	var (
		clusterID = "test-uid"

		kubeNamespace = &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: metav1.NamespaceSystem,
				UID:  "test-uid",
			},
		}

		nodeList = &v1.NodeList{
			Items: []v1.Node{{}},
		}
	)

	It("collects the deployment context", func() {
		collector := licensing.NewDeploymentContextCollector(licensing.DeploymentContextCollectorConfig{
			K8sClientReader: fake.NewFakeClient(kubeNamespace, nodeList),
			PodUID:          "pod-uid",
		})

		expCtx := dataplane.DeploymentContext{
			Integration:      "ngf",
			ClusterID:        clusterID,
			InstallationID:   "pod-uid",
			ClusterNodeCount: 1,
		}

		depCtx, err := collector.Collect(context.Background())
		Expect(err).ToNot(HaveOccurred())
		Expect(depCtx).To(Equal(expCtx))
	})

	It("returns an error and default deployment context if cluster info isn't found", func() {
		collector := licensing.NewDeploymentContextCollector(licensing.DeploymentContextCollectorConfig{
			K8sClientReader: fake.NewFakeClient(),
			PodUID:          "pod-uid",
		})

		expCtx := dataplane.DeploymentContext{
			Integration:    "ngf",
			InstallationID: "pod-uid",
		}

		depCtx, err := collector.Collect(context.Background())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("error collecting cluster ID and cluster node count"))
		Expect(depCtx).To(Equal(expCtx))
	})
})
