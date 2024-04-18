package suite

import (
	"context"
	"errors"
	"net/http"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/nginxinc/nginx-gateway-fabric/tests/framework"
)

const (
	// FIXME(bjee19): Find an automated way to keep the version updated here similar to dependabot.
	// https://github.com/nginxinc/nginx-gateway-fabric/issues/1665
	debugImage         = "busybox:1.28"
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

	BeforeAll(func() {
		cfg := getDefaultSetupCfg()
		cfg.nfr = true
		setup(cfg, "--set", "nginxGateway.securityContext.runAsNonRoot=false")
	})

	BeforeEach(func() {
		Expect(resourceManager.Apply([]client.Object{ns})).To(Succeed())
		Expect(resourceManager.ApplyFromFiles(files, ns.Name)).To(Succeed())
		Expect(resourceManager.WaitForAppsToBeReady(ns.Name)).To(Succeed())

		err := waitForWorkingTraffic()
		Expect(err).ToNot(HaveOccurred())
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

		checkContainerLogsForErrors(podNames[0])

		err = waitForWorkingTraffic()
		Expect(err).ToNot(HaveOccurred())

		// I tried just deleting the routes and ran into a bunch of issues, deleting all the files was better
		Expect(resourceManager.DeleteFromFiles(files, ns.Name)).To(Succeed())

		err = waitForFailingTraffic()
		Expect(err).ToNot(HaveOccurred())

		Expect(resourceManager.ApplyFromFiles(files, ns.Name)).To(Succeed())
		Expect(resourceManager.WaitForAppsToBeReady(ns.Name)).To(Succeed())

		err = waitForWorkingTraffic()
		Expect(err).ToNot(HaveOccurred())
	})

	It("recovers when nginx container is restarted", func() {
		podNames, err := framework.GetReadyNGFPodNames(k8sClient, ngfNamespace, releaseName, timeoutConfig.GetTimeout)
		Expect(err).ToNot(HaveOccurred())
		Expect(podNames).ToNot(BeEmpty())

		output, err := restartNginxContainer(nginxContainerName)
		Expect(err).ToNot(HaveOccurred(), string(output))

		checkContainerLogsForErrors(podNames[0])

		err = waitForWorkingTraffic()
		Expect(err).ToNot(HaveOccurred())

		Expect(resourceManager.DeleteFromFiles(files, ns.Name)).To(Succeed())

		err = waitForFailingTraffic()
		Expect(err).ToNot(HaveOccurred())

		Expect(resourceManager.ApplyFromFiles(files, ns.Name)).To(Succeed())
		Expect(resourceManager.WaitForAppsToBeReady(ns.Name)).To(Succeed())

		err = waitForWorkingTraffic()
		Expect(err).ToNot(HaveOccurred())
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

	err = waitForContainerRestart(podNames[0], nginxContainerName, restartCount)
	Expect(err).ToNot(HaveOccurred())

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
		"--image="+debugImage,
		"--target=nginx-gateway",
		"--",
		"sh",
		"-c",
		"$(PID=$(pgrep -f \"/[u]sr/bin/gateway\") && kill -9 $PID)").CombinedOutput()
	if err != nil {
		return output, err
	}

	err = waitForContainerRestart(podNames[0], ngfContainerName, restartCount)
	Expect(err).ToNot(HaveOccurred())

	return nil, nil
}

func waitForContainerRestart(ngfPodName string, containerName string, currentRestartCount int) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutConfig.RequestTimeout)
	defer cancel()

	//nolint:nilerr
	return wait.PollUntilContextCancel(
		ctx,
		500*time.Millisecond,
		true, /* poll immediately */
		func(ctx context.Context) (bool, error) {
			var ngfPod core.Pod
			if err := k8sClient.Get(ctx, types.NamespacedName{Namespace: ngfNamespace, Name: ngfPodName}, &ngfPod); err != nil {
				return false, nil
			}

			for _, containerStatus := range ngfPod.Status.ContainerStatuses {
				if containerStatus.Name == containerName {
					return int(containerStatus.RestartCount) == currentRestartCount+1, nil
				}
			}
			return false, nil
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
			if err := expectRequest(teaURL, address, http.StatusOK, "URI: /tea"); err != nil {
				return false, nil
			}
			if err := expectRequest(coffeeURL, address, http.StatusOK, "URI: /coffee"); err != nil {
				return false, nil
			}
			return true, nil
		},
	)
}

func waitForFailingTraffic() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutConfig.RequestTimeout)
	defer cancel()

	return wait.PollUntilContextCancel(
		ctx,
		500*time.Millisecond,
		true, /* poll immediately */
		func(_ context.Context) (bool, error) {
			if err := expectRequest(teaURL, address, 0, "URI: /tea"); err == nil {
				return false, nil
			}
			if err := expectRequest(coffeeURL, address, 0, "URI: /coffee"); err == nil {
				return false, nil
			}
			return true, nil
		},
	)
}

func expectRequest(appURL string, address string, httpStatus int, responseBodyMessage string) error {
	status, body, err := framework.Get(appURL, address, timeoutConfig.RequestTimeout)
	if status != httpStatus {
		return errors.New("http statuses were not equal")
	}
	if httpStatus == http.StatusOK {
		if !strings.Contains(body, responseBodyMessage) {
			return errors.New("expected response body to contain body message")
		}
	} else {
		if strings.Contains(body, responseBodyMessage) {
			return errors.New("expected response body to not contain body message")
		}
	}
	return err
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
