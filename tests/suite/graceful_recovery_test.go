package suite

import (
	"context"
	"net/http"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/nginxinc/nginx-gateway-fabric/tests/framework"
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

	nginxContainerName := "nginx"
	ngfContainerName := "nginx-gateway"
	teaURL := "https://cafe.example.com/tea"
	coffeeURL := "http://cafe.example.com/coffee"

	BeforeAll(func() {
		cfg := getDefaultSetupCfg()
		cfg.nfr = true
		setup(cfg, "--set", "nginxGateway.securityContext.runAsNonRoot=false")
	})

	BeforeEach(func() {
		Expect(resourceManager.Apply([]client.Object{ns})).To(Succeed())
		Expect(resourceManager.ApplyFromFiles(files, ns.Name)).To(Succeed())
		Expect(resourceManager.WaitForAppsToBeReady(ns.Name)).To(Succeed())
		// Sometimes the traffic would error with code 502, after implementing this sleep it stopped.
		time.Sleep(2 * time.Second)

		expectWorkingTraffic(teaURL, coffeeURL)
	})

	AfterAll(func() {
		Expect(resourceManager.DeleteFromFiles(files, ns.Name)).To(Succeed())
		Expect(resourceManager.Delete([]client.Object{ns})).To(Succeed())
	})

	It("recovers when NGF container is restarted", func() {
		podNames, err := framework.GetReadyNGFPodNames(k8sClient, ngfNamespace, releaseName, timeoutConfig.GetTimeout)
		Expect(err).ToNot(HaveOccurred())
		Expect(podNames).ToNot(BeEmpty())

		output, err := restartNGFProcess(ngfContainerName)
		Expect(err).ToNot(HaveOccurred(), string(output))

		expectWorkingTraffic(teaURL, coffeeURL)

		checkContainerLogsForErrors(podNames[0])

		// I tried just deleting the routes and ran into a bunch of issues, deleting all the files was better
		Expect(resourceManager.DeleteFromFiles(files, ns.Name)).To(Succeed())
		// Wait for files to be deleted.
		time.Sleep(2 * time.Second)

		expectFailingTraffic(teaURL, coffeeURL)

		Expect(resourceManager.ApplyFromFiles(files, ns.Name)).To(Succeed())
		Expect(resourceManager.WaitForAppsToBeReady(ns.Name)).To(Succeed())

		expectWorkingTraffic(teaURL, coffeeURL)
	})

	It("recovers when nginx container is restarted", func() {
		podNames, err := framework.GetReadyNGFPodNames(k8sClient, ngfNamespace, releaseName, timeoutConfig.GetTimeout)
		Expect(err).ToNot(HaveOccurred())
		Expect(podNames).ToNot(BeEmpty())

		output, err := restartNginxContainer(nginxContainerName)
		Expect(err).ToNot(HaveOccurred(), string(output))

		checkContainerLogsForErrors(podNames[0])

		Expect(resourceManager.DeleteFromFiles(files, ns.Name)).To(Succeed())
		// Wait for files to be deleted.
		time.Sleep(2 * time.Second)

		expectFailingTraffic(teaURL, coffeeURL)

		Expect(resourceManager.ApplyFromFiles(files, ns.Name)).To(Succeed())
		Expect(resourceManager.WaitForAppsToBeReady(ns.Name)).To(Succeed())

		expectWorkingTraffic(teaURL, coffeeURL)
	})
})

func restartNginxContainer(nginxContainerName string) ([]byte, error) {
	podNames, err := framework.GetReadyNGFPodNames(k8sClient, ngfNamespace, releaseName, timeoutConfig.GetTimeout)
	Expect(err).ToNot(HaveOccurred())
	Expect(podNames).ToNot(BeEmpty())

	ctx, cancel := context.WithTimeout(context.Background(), timeoutConfig.GetTimeout)
	defer cancel()

	var ngfPod core.Pod
	err = k8sClient.Get(ctx, types.NamespacedName{Namespace: ngfNamespace, Name: podNames[0]}, &ngfPod)
	Expect(err).ToNot(HaveOccurred())

	var restartCount int
	for _, containerStatus := range ngfPod.Status.ContainerStatuses {
		if containerStatus.Name == nginxContainerName {
			restartCount = int(containerStatus.RestartCount)
		}
	}

	output, err := exec.Command( // nolint:gosec
		"kubectl",
		"exec",
		"-n",
		ngfNamespace,
		podNames[0],
		"--container",
		"nginx",
		"--",
		"sh",
		"-c",
		"$(PID=$(pgrep -f \"[n]ginx: master process\") && kill -9 $PID)").CombinedOutput()
	if err != nil {
		return output, err
	}

	// Wait for NGF to restart.
	time.Sleep(2 * time.Second)

	err = k8sClient.Get(ctx, types.NamespacedName{Namespace: ngfNamespace, Name: podNames[0]}, &ngfPod)
	Expect(err).ToNot(HaveOccurred())

	for _, containerStatus := range ngfPod.Status.ContainerStatuses {
		if containerStatus.Name == nginxContainerName {
			Expect(int(containerStatus.RestartCount)).To(Equal(restartCount + 1))
		}
	}
	return nil, nil
}

func restartNGFProcess(ngfContainerName string) ([]byte, error) {
	podNames, err := framework.GetReadyNGFPodNames(k8sClient, ngfNamespace, releaseName, timeoutConfig.GetTimeout)
	Expect(err).ToNot(HaveOccurred())
	Expect(podNames).ToNot(BeEmpty())

	ctx, cancel := context.WithTimeout(context.Background(), timeoutConfig.GetTimeout)
	defer cancel()

	var ngfPod core.Pod
	err = k8sClient.Get(ctx, types.NamespacedName{Namespace: ngfNamespace, Name: podNames[0]}, &ngfPod)
	Expect(err).ToNot(HaveOccurred())

	var restartCount int
	for _, containerStatus := range ngfPod.Status.ContainerStatuses {
		if containerStatus.Name == ngfContainerName {
			restartCount = int(containerStatus.RestartCount)
		}
	}

	output, err := exec.Command( // nolint:gosec
		"kubectl",
		"debug",
		"-n",
		ngfNamespace,
		podNames[0],
		"--image=busybox:1.28",
		"--target=nginx-gateway",
		"--",
		"sh",
		"-c",
		"$(PID=$(pgrep -f \"/[u]sr/bin/gateway\") && kill -9 $PID)").CombinedOutput()
	if err != nil {
		return output, err
	}

	// Wait for NGF to restart.
	time.Sleep(6 * time.Second)

	err = k8sClient.Get(ctx, types.NamespacedName{Namespace: ngfNamespace, Name: podNames[0]}, &ngfPod)
	Expect(err).ToNot(HaveOccurred())

	for _, containerStatus := range ngfPod.Status.ContainerStatuses {
		if containerStatus.Name == ngfContainerName {
			Expect(int(containerStatus.RestartCount)).To(Equal(restartCount + 1))
		}
	}
	return nil, nil
}

func expectWorkingTraffic(teaURL, coffeeURL string) {
	status, body, err := framework.Get(teaURL, address, timeoutConfig.RequestTimeout)
	Expect(err).ToNot(HaveOccurred())
	Expect(status).To(Equal(http.StatusOK))
	Expect(body).To(ContainSubstring("URI: /tea"))

	status, body, err = framework.Get(coffeeURL, address, timeoutConfig.RequestTimeout)
	Expect(err).ToNot(HaveOccurred())
	Expect(status).To(Equal(http.StatusOK), coffeeURL+" "+address)
	Expect(body).To(ContainSubstring("URI: /coffee"))
}

func expectFailingTraffic(teaURL, coffeeURL string) {
	status, body, err := framework.Get(teaURL, address, timeoutConfig.RequestTimeout)
	Expect(err).To(HaveOccurred())
	Expect(status).ToNot(Equal(http.StatusOK))
	Expect(body).ToNot(ContainSubstring("URI: /tea"))

	status, body, err = framework.Get(coffeeURL, address, timeoutConfig.RequestTimeout)
	Expect(err).To(HaveOccurred())
	Expect(status).ToNot(Equal(http.StatusOK))
	Expect(body).ToNot(ContainSubstring("URI: /coffee"))
}

func checkContainerLogsForErrors(ngfPodName string) {
	sinceSeconds := int64(15)
	logs, err := resourceManager.GetPodLogs(
		ngfNamespace,
		ngfPodName,
		&core.PodLogOptions{Container: "nginx", SinceSeconds: &sinceSeconds},
	)
	Expect(err).ToNot(HaveOccurred())
	Expect(logs).ToNot(ContainSubstring("emerg"), logs)

	logs, err = resourceManager.GetPodLogs(
		ngfNamespace,
		ngfPodName,
		&core.PodLogOptions{Container: "nginx-gateway", SinceSeconds: &sinceSeconds},
	)
	Expect(err).ToNot(HaveOccurred())
	Expect(logs).ToNot(ContainSubstring("error"), logs)
}
