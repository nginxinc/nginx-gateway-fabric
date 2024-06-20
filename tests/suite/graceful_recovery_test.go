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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/nginxinc/nginx-gateway-fabric/tests/framework"
)

const (
	nginxContainerName = "nginx"
	ngfContainerName   = "nginx-gateway"
)

// Since checkContainerLogsForErrors may experience interference from previous tests (as explained in the function
// documentation), this test is recommended to be run separate from other tests.
var _ = Describe("Graceful Recovery test", Ordered, Label("functional", "graceful-recovery"), func() {
	files := []string{
		"graceful-recovery/cafe.yaml",
		"graceful-recovery/cafe-secret.yaml",
		"graceful-recovery/gateway.yaml",
		"graceful-recovery/cafe-routes.yaml",
	}

	var ns core.Namespace

	baseHTTPURL := "http://cafe.example.com"
	baseHTTPSURL := "https://cafe.example.com"
	teaURL := baseHTTPSURL + "/tea"
	coffeeURL := baseHTTPURL + "/coffee"

	var ngfPodName string

	BeforeAll(func() {
		// this test is unique in that it will check the entire log of both ngf and nginx containers
		// for any errors, so in order to avoid errors generated in previous tests we will uninstall
		// NGF installed at the suite level, then re-deploy our own
		teardown(releaseName)

		setup(getDefaultSetupCfg())

		podNames, err := framework.GetReadyNGFPodNames(k8sClient, ngfNamespace, releaseName, timeoutConfig.GetTimeout)
		Expect(err).ToNot(HaveOccurred())
		Expect(podNames).To(HaveLen(1))

		ngfPodName = podNames[0]
		if portFwdPort != 0 {
			coffeeURL = fmt.Sprintf("%s:%d/coffee", baseHTTPURL, portFwdPort)
		}
		if portFwdHTTPSPort != 0 {
			teaURL = fmt.Sprintf("%s:%d/tea", baseHTTPSURL, portFwdHTTPSPort)
		}
	})

	BeforeEach(func() {
		ns = core.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "graceful-recovery",
			},
		}

		Expect(resourceManager.Apply([]client.Object{&ns})).To(Succeed())
		Expect(resourceManager.ApplyFromFiles(files, ns.Name)).To(Succeed())
		Expect(resourceManager.WaitForAppsToBeReadyWithPodCount(ns.Name, 2)).To(Succeed())

		Eventually(
			func() error {
				return checkForWorkingTraffic(teaURL, coffeeURL)
			}).
			WithTimeout(timeoutConfig.RequestTimeout).
			WithPolling(500 * time.Millisecond).
			Should(Succeed())
	})

	AfterAll(func() {
		Expect(resourceManager.DeleteFromFiles(files, ns.Name)).To(Succeed())
		Expect(resourceManager.DeleteNamespace(ns.Name)).To(Succeed())
	})

	It("recovers when NGF container is restarted", func() {
		runRecoveryTest(teaURL, coffeeURL, ngfPodName, ngfContainerName, files, &ns)
	})

	It("recovers when nginx container is restarted", func() {
		runRecoveryTest(teaURL, coffeeURL, ngfPodName, nginxContainerName, files, &ns)
	})
})

func runRecoveryTest(teaURL, coffeeURL, ngfPodName, containerName string, files []string, ns *core.Namespace) {
	var (
		err       error
		leaseName string
	)

	if containerName != nginxContainerName {
		// Since we have already deployed resources and ran resourceManager.WaitForAppsToBeReadyWithPodCount earlier,
		// we know that the applications are ready at this point. This could only be the case if NGF has written
		// statuses, which could only be the case if NGF has the leader lease. Since there is only one instance
		// of NGF in this test, we can be certain that this is the correct leaseholder name.
		leaseName, err = getLeaderElectionLeaseHolderName()
		Expect(err).ToNot(HaveOccurred())
	}

	restartContainer(ngfPodName, containerName)

	if containerName != nginxContainerName {
		Eventually(
			func() error {
				return checkLeaderLeaseChange(leaseName)
			}).
			WithTimeout(timeoutConfig.GetLeaderLeaseTimeout).
			WithPolling(500 * time.Millisecond).
			Should(Succeed())
	}

	// TEMP: Add sleep to see if it resolves falures
	time.Sleep(30 * time.Second)

	Eventually(
		func() error {
			return checkForWorkingTraffic(teaURL, coffeeURL)
		}).
		WithTimeout(timeoutConfig.RequestTimeout).
		WithPolling(500 * time.Millisecond).
		Should(Succeed())

	Expect(resourceManager.DeleteFromFiles(files, ns.Name)).To(Succeed())

	Eventually(
		func() error {
			return checkForFailingTraffic(teaURL, coffeeURL)
		}).
		WithTimeout(timeoutConfig.RequestTimeout).
		WithPolling(500 * time.Millisecond).
		Should(Succeed())

	Expect(resourceManager.ApplyFromFiles(files, ns.Name)).To(Succeed())
	Expect(resourceManager.WaitForAppsToBeReadyWithPodCount(ns.Name, 2)).To(Succeed())

	Eventually(
		func() error {
			return checkForWorkingTraffic(teaURL, coffeeURL)
		}).
		WithTimeout(timeoutConfig.RequestTimeout).
		WithPolling(500 * time.Millisecond).
		Should(Succeed())

	checkContainerLogsForErrors(ngfPodName)
}

func restartContainer(ngfPodName, containerName string) {
	var jobScript string
	if containerName == "nginx" {
		jobScript = "PID=$(pgrep -f \"nginx: master process\") && kill -9 $PID"
	} else {
		jobScript = "PID=$(pgrep -f \"/usr/bin/gateway\") && kill -9 $PID"
	}

	restartCount, err := getContainerRestartCount(ngfPodName, containerName)
	Expect(err).ToNot(HaveOccurred())

	job, err := runNodeDebuggerJob(ngfPodName, jobScript)
	Expect(err).ToNot(HaveOccurred())

	Eventually(
		func() error {
			return checkContainerRestart(ngfPodName, containerName, restartCount)
		}).
		WithTimeout(timeoutConfig.ContainerRestartTimeout).
		WithPolling(500 * time.Millisecond).
		Should(Succeed())

	// default propagation policy is metav1.DeletePropagationOrphan which does not delete the underlying
	// pod created through the job after the job is deleted. Setting it to metav1.DeletePropagationBackground
	// deletes the underlying pod after the job is deleted.
	Expect(resourceManager.Delete(
		[]client.Object{job},
		client.PropagationPolicy(metav1.DeletePropagationBackground),
	)).To(Succeed())
}

func checkContainerRestart(ngfPodName, containerName string, currentRestartCount int) error {
	restartCount, err := getContainerRestartCount(ngfPodName, containerName)
	if err != nil {
		return err
	}

	if restartCount != currentRestartCount+1 {
		return fmt.Errorf("expected current restart count: %d to match incremented restart count: %d",
			restartCount, currentRestartCount+1)
	}

	return nil
}

func checkForWorkingTraffic(teaURL, coffeeURL string) error {
	if err := expectRequestToSucceed(teaURL, address, "URI: /tea"); err != nil {
		return err
	}
	if err := expectRequestToSucceed(coffeeURL, address, "URI: /coffee"); err != nil {
		return err
	}
	return nil
}

func checkForFailingTraffic(teaURL, coffeeURL string) error {
	if err := expectRequestToFail(teaURL, address); err != nil {
		return err
	}
	if err := expectRequestToFail(coffeeURL, address); err != nil {
		return err
	}
	return nil
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

func expectRequestToFail(appURL, address string) error {
	status, body, err := framework.Get(appURL, address, timeoutConfig.RequestTimeout)
	if status != 0 {
		return errors.New("expected http status to be 0")
	}

	if body != "" {
		return fmt.Errorf("expected response body to be empty, instead received: %s", body)
	}

	if err == nil {
		return errors.New("expected request to error")
	}

	return nil
}

// checkContainerLogsForErrors checks both nginx and ngf container's logs for any possible errors.
// Since this function retrieves all the logs from both containers and the NGF pod may be shared between tests,
// the logs retrieved may contain log messages from previous tests, thus any errors in the logs from previous tests
// may cause an interference with this test and cause this test to fail.
func checkContainerLogsForErrors(ngfPodName string) {
	logs, err := resourceManager.GetPodLogs(
		ngfNamespace,
		ngfPodName,
		&core.PodLogOptions{Container: nginxContainerName},
	)
	Expect(err).ToNot(HaveOccurred())

	for _, line := range strings.Split(logs, "\n") {
		Expect(line).ToNot(ContainSubstring("[crit]"), line)
		Expect(line).ToNot(ContainSubstring("[alert]"), line)
		Expect(line).ToNot(ContainSubstring("[emerg]"), line)
		if strings.Contains(line, "[error]") {
			expectedError1 := "connect() failed (111: Connection refused)"
			// FIXME(salonichf5) remove this error message check
			// when https://github.com/nginxinc/nginx-gateway-fabric/issues/2090 is completed.
			expectedError2 := "no live upstreams while connecting to upstream"
			Expect(line).To(Or(ContainSubstring(expectedError1), ContainSubstring(expectedError2)))
		}
	}

	logs, err = resourceManager.GetPodLogs(
		ngfNamespace,
		ngfPodName,
		&core.PodLogOptions{Container: ngfContainerName},
	)
	Expect(err).ToNot(HaveOccurred())

	for _, line := range strings.Split(logs, "\n") {
		if *plusEnabled && strings.Contains(line, "\"level\":\"error\"") {
			Expect(line).To(ContainSubstring("Usage reporting must be enabled when using NGINX Plus"), line)
		} else {
			Expect(line).ToNot(ContainSubstring("\"level\":\"error\""), line)
		}
	}
}

func checkLeaderLeaseChange(originalLeaseName string) error {
	leaseName, err := getLeaderElectionLeaseHolderName()
	if err != nil {
		return err
	}

	if originalLeaseName == leaseName {
		return fmt.Errorf("expected originalLeaseName: %s, to not match current leaseName: %s", originalLeaseName, leaseName)
	}

	return nil
}

func getLeaderElectionLeaseHolderName() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutConfig.GetTimeout)
	defer cancel()

	var lease coordination.Lease
	key := types.NamespacedName{Name: "ngf-test-nginx-gateway-fabric-leader-election", Namespace: ngfNamespace}

	if err := k8sClient.Get(ctx, key, &lease); err != nil {
		return "", errors.New("could not retrieve leader election lease")
	}

	if *lease.Spec.HolderIdentity == "" {
		return "", errors.New("leader election lease holder identity is empty")
	}

	return *lease.Spec.HolderIdentity, nil
}

func getContainerRestartCount(ngfPodName, containerName string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutConfig.GetTimeout)
	defer cancel()

	var ngfPod core.Pod
	if err := k8sClient.Get(ctx, types.NamespacedName{Namespace: ngfNamespace, Name: ngfPodName}, &ngfPod); err != nil {
		return 0, fmt.Errorf("error retrieving NGF Pod: %w", err)
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
		return nil, fmt.Errorf("error retrieving NGF Pod: %w", err)
	}

	b, err := resourceManager.GetFileContents("graceful-recovery/node-debugger-job.yaml")
	if err != nil {
		return nil, fmt.Errorf("error processing node debugger job file: %w", err)
	}

	job := &v1.Job{}
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
