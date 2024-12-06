package licensing_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/licensing"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
)

var _ = Describe("DeploymentContextCollector", func() {
	var (
		clusterID = "test-uid"
		ngfPod    = &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pod1",
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind: "ReplicaSet",
						Name: "replicaset1",
					},
				},
			},
		}

		ngfReplicaSet = &appsv1.ReplicaSet{
			Spec: appsv1.ReplicaSetSpec{
				Replicas: helpers.GetPointer[int32](1),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "replicaset1",
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind: "Deployment",
						Name: "Deployment1",
						UID:  "test-uid-replicaSet",
					},
				},
			},
		}

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
			K8sClientReader: fake.NewFakeClient(ngfPod, ngfReplicaSet, kubeNamespace, nodeList),
			PodNSName:       types.NamespacedName{Name: ngfPod.Name},
		})

		expCtx := dataplane.DeploymentContext{
			Integration:      "ngf",
			ClusterID:        clusterID,
			InstallationID:   "test-uid-replicaSet",
			ClusterNodeCount: 1,
		}

		depCtx, err := collector.Collect(context.Background())
		Expect(err).ToNot(HaveOccurred())
		Expect(depCtx).To(Equal(expCtx))
	})

	It("returns an error if cluster info isn't found", func() {
		collector := licensing.NewDeploymentContextCollector(licensing.DeploymentContextCollectorConfig{
			K8sClientReader: fake.NewFakeClient(),
		})

		_, err := collector.Collect(context.Background())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("error getting cluster information"))
	})

	It("returns the deployment context when the replicaset isn't found", func() {
		collector := licensing.NewDeploymentContextCollector(licensing.DeploymentContextCollectorConfig{
			K8sClientReader: fake.NewFakeClient(ngfPod, kubeNamespace, nodeList),
			PodNSName:       types.NamespacedName{Name: ngfPod.Name},
		})

		expCtx := dataplane.DeploymentContext{
			Integration:      "ngf",
			ClusterID:        clusterID,
			ClusterNodeCount: 1,
		}

		depCtx, err := collector.Collect(context.Background())
		Expect(err).ToNot(HaveOccurred())
		Expect(depCtx).To(Equal(expCtx))
	})

	It("returns the deployment context when the replicaset doesn't have a uid", func() {
		ngfReplicaSet.ObjectMeta.OwnerReferences[0].UID = ""

		collector := licensing.NewDeploymentContextCollector(licensing.DeploymentContextCollectorConfig{
			K8sClientReader: fake.NewFakeClient(ngfPod, ngfReplicaSet, kubeNamespace, nodeList),
			PodNSName:       types.NamespacedName{Name: ngfPod.Name},
		})

		expCtx := dataplane.DeploymentContext{
			Integration:      "ngf",
			ClusterID:        clusterID,
			ClusterNodeCount: 1,
		}

		depCtx, err := collector.Collect(context.Background())
		Expect(err).ToNot(HaveOccurred())
		Expect(depCtx).To(Equal(expCtx))
	})
})
