package suite

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/batch/v1"
	coordination "k8s.io/api/coordination/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/nginxinc/nginx-gateway-fabric/tests/framework"
)

const (
	teaURL             = "https://cafe.example.com/tea"
	coffeeURL          = "http://cafe.example.com/coffee"
	nginxContainerName = "nginx"
	ngfContainerName   = "nginx-gateway"
)

var _ = Describe("Graceful Recovery test", Ordered, Label("nfr", "graceful-recovery"), func() {
	files := []string{
		"graceful-recovery/cafe.yaml",
		"graceful-recovery/cafe-secret.yaml",
		"graceful-recovery/gateway.yaml",
		"graceful-recovery/cafe-routes.yaml",
	}
	ns := &core.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "graceful-recovery",
		},
	}

	BeforeEach(func() {
		Expect(resourceManager.Apply([]client.Object{ns})).To(Succeed())
		Expect(resourceManager.ApplyFromFiles(files, ns.Name)).To(Succeed())
		Expect(resourceManager.WaitForAppsToBeReady(ns.Name)).To(Succeed())

		Expect(waitForWorkingTraffic()).ToNot(HaveOccurred())
	})

	AfterAll(func() {
		Expect(resourceManager.DeleteFromFiles(files, ns.Name)).To(Succeed())
		Expect(resourceManager.Delete([]client.Object{ns})).To(Succeed())
	})

	It("recovers when NGF container is restarted", func() {
		runRecoveryTest(ngfContainerName, files, ns)
	})

	It("recovers when nginx container is restarted", func() {
		runRecoveryTest(nginxContainerName, files, ns)
	})
})

func runRecoveryTest(containerName string, files []string, ns *core.Namespace) {
	podNames, err := framework.GetReadyNGFPodNames(k8sClient, ngfNamespace, releaseName, timeoutConfig.GetTimeout)
	Expect(err).ToNot(HaveOccurred())
	Expect(podNames).ToNot(BeEmpty())

	leaseName, err := getLeaderElectionLeaseHolderName()
	Expect(err).ToNot(HaveOccurred())

	restartContainer(containerName)

	checkContainerLogsForErrors(podNames[0])

	Expect(waitForLeaderLeaseToChange(leaseName)).ToNot(HaveOccurred())

	Expect(waitForWorkingTraffic()).ToNot(HaveOccurred())

	Expect(resourceManager.DeleteFromFiles(files, ns.Name)).To(Succeed())

	Expect(waitForFailingTraffic()).ToNot(HaveOccurred())

	Expect(resourceManager.ApplyFromFiles(files, ns.Name)).To(Succeed())
	Expect(resourceManager.WaitForAppsToBeReady(ns.Name)).To(Succeed())

	Expect(waitForWorkingTraffic()).ToNot(HaveOccurred())
}

func restartContainer(containerName string) {
	var jobScript string
	if containerName == "nginx" {
		jobScript = "PID=$(pgrep -f \"nginx: master process\") && kill -9 $PID"
	} else {
		jobScript = "PID=$(pgrep -f \"/usr/bin/gateway\") && kill -9 $PID"
	}

	podNames, err := framework.GetReadyNGFPodNames(k8sClient, ngfNamespace, releaseName, timeoutConfig.GetTimeout)
	Expect(err).ToNot(HaveOccurred())
	Expect(podNames).ToNot(BeEmpty())

	restartCount, err := getContainerRestartCount(containerName, podNames[0])
	Expect(err).ToNot(HaveOccurred())

	job, err := runNodeDebuggerJob(podNames[0], jobScript)
	Expect(err).ToNot(HaveOccurred())

	Expect(waitForContainerRestart(podNames[0], containerName, restartCount)).ToNot(HaveOccurred())

	// propagation policy is set to delete underlying pod created through job
	Expect(resourceManager.Delete(
		[]client.Object{job},
		client.PropagationPolicy(metav1.DeletePropagationBackground),
	)).ToNot(HaveOccurred())
}

func waitForContainerRestart(ngfPodName, containerName string, currentRestartCount int) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutConfig.RequestTimeout)
	defer cancel()

	//nolint:nilerr
	return wait.PollUntilContextCancel(
		ctx,
		500*time.Millisecond,
		true, /* poll immediately */
		func(_ context.Context) (bool, error) {
			restartCount, err := getContainerRestartCount(containerName, ngfPodName)
			if err != nil {
				return false, nil
			}

			return restartCount == currentRestartCount+1, nil
		},
	)
}

func waitForWorkingTraffic() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutConfig.RequestTimeout)
	defer cancel()

	//nolint:nilerr
	return wait.PollUntilContextCancel(
		ctx,
		500*time.Millisecond,
		true, /* poll immediately */
		func(_ context.Context) (bool, error) {
			if err := expectRequestToSucceed(teaURL, address, "URI: /tea"); err != nil {
				return false, nil
			}
			if err := expectRequestToSucceed(coffeeURL, address, "URI: /coffee"); err != nil {
				return false, nil
			}
			return true, nil
		},
	)
}

func waitForFailingTraffic() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutConfig.RequestTimeout)
	defer cancel()

	//nolint:nilerr
	return wait.PollUntilContextCancel(
		ctx,
		500*time.Millisecond,
		true, /* poll immediately */
		func(_ context.Context) (bool, error) {
			if err := expectRequestToFail(teaURL, address, "URI: /tea"); err != nil {
				return false, nil
			}
			if err := expectRequestToFail(coffeeURL, address, "URI: /coffee"); err != nil {
				return false, nil
			}
			return true, nil
		},
	)
}

func expectRequestToSucceed(appURL, address string, responseBodyMessage string) error {
	status, body, err := framework.Get(appURL, address, timeoutConfig.RequestTimeout)
	if status != http.StatusOK {
		return errors.New("http status was not 200")
	}

	if !strings.Contains(body, responseBodyMessage) {
		return errors.New("expected response body to contain correct body message")
	}

	return err
}

func expectRequestToFail(appURL, address string, responseBodyMessage string) error {
	status, body, err := framework.Get(appURL, address, timeoutConfig.RequestTimeout)
	if status != 0 {
		return errors.New("expected http status to be 0")
	}

	if strings.Contains(body, responseBodyMessage) {
		return errors.New("expected response body to not contain correct body message")
	}

	if err == nil {
		return errors.New("expected request to error")
	}

	return nil
}

func checkContainerLogsForErrors(ngfPodName string) {
	sinceSeconds := int64(15)
	logs, err := resourceManager.GetPodLogs(
		ngfNamespace,
		ngfPodName,
		&core.PodLogOptions{Container: nginxContainerName, SinceSeconds: &sinceSeconds},
	)
	Expect(err).ToNot(HaveOccurred())
	Expect(logs).ToNot(ContainSubstring("emerg"), logs)

	logs, err = resourceManager.GetPodLogs(
		ngfNamespace,
		ngfPodName,
		&core.PodLogOptions{Container: ngfContainerName, SinceSeconds: &sinceSeconds},
	)
	Expect(err).ToNot(HaveOccurred())
	Expect(logs).ToNot(ContainSubstring("error"), logs)
}

func waitForLeaderLeaseToChange(originalLeaseName string) error {
	leaseCtx, leaseCancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer leaseCancel()

	//nolint:nilerr
	return wait.PollUntilContextCancel(
		leaseCtx,
		500*time.Millisecond,
		true, /* poll immediately */
		func(_ context.Context) (bool, error) {
			leaseName, err := getLeaderElectionLeaseHolderName()
			if err != nil {
				return false, nil
			}

			if originalLeaseName != leaseName {
				return true, nil
			}

			return false, nil
		},
	)
}

func getLeaderElectionLeaseHolderName() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutConfig.GetTimeout)
	defer cancel()

	var lease coordination.Lease
	key := types.NamespacedName{Name: "ngf-test-nginx-gateway-fabric-leader-election", Namespace: ngfNamespace}

	if err := k8sClient.Get(ctx, key, &lease); err != nil {
		return "", errors.New("could not retrieve leader election lease")
	}
	return *lease.Spec.HolderIdentity, nil
}

func getContainerRestartCount(containerName, ngfPodName string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutConfig.GetTimeout)
	defer cancel()

	var ngfPod core.Pod
	if err := k8sClient.Get(ctx, types.NamespacedName{Namespace: ngfNamespace, Name: ngfPodName}, &ngfPod); err != nil {
		return 0, fmt.Errorf("error retriving NGF Pod: %w", err)
	}

	var restartCount int
	for _, containerStatus := range ngfPod.Status.ContainerStatuses {
		if containerStatus.Name == containerName {
			restartCount = int(containerStatus.RestartCount)
		}
	}

	return restartCount, nil
}

func runNodeDebuggerJob(ngfPodName, jobScript string) (*v1.Job, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutConfig.GetTimeout)
	defer cancel()

	var ngfPod core.Pod
	if err := k8sClient.Get(ctx, types.NamespacedName{Namespace: ngfNamespace, Name: ngfPodName}, &ngfPod); err != nil {
		return nil, fmt.Errorf("error retriving NGF Pod: %w", err)
	}

	b, err := resourceManager.GetFileContents("graceful-recovery/node-debugger-job.yaml")
	if err != nil {
		return nil, fmt.Errorf("error processing node debugger job file: %w", err)
	}

	job := &v1.Job{}
	_ = v1.AddToScheme(resourceManager.K8sClient.Scheme())
	if err = yaml.Unmarshal(b.Bytes(), job); err != nil {
		return nil, fmt.Errorf("error with yaml unmarshal: %w", err)
	}

	job.Spec.Template.Spec.NodeSelector["kubernetes.io/hostname"] = ngfPod.Spec.NodeName
	if len(job.Spec.Template.Spec.Containers) != 1 {
		return nil, fmt.Errorf(
			"expected node debugger job to contain one container, actual number: %d",
			len(job.Spec.Template.Spec.Containers),
		)
	}
	job.Spec.Template.Spec.Containers[0].Args = []string{jobScript}
	job.Namespace = ngfNamespace

	if err = resourceManager.Apply([]client.Object{job}); err != nil {
		return nil, fmt.Errorf("error in applying job: %w", err)
	}

	return job, nil
}
