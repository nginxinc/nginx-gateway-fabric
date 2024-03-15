package suite

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	coordination "k8s.io/api/coordination/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	v1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/nginxinc/nginx-gateway-fabric/tests/framework"
)

// This test installs the latest released version of NGF, then upgrades to the edge version (or dev version).
// During the upgrade, traffic is continuously sent to ensure no downtime.
// We also check that the leader election lease has been updated, and that Gateway updates are processed.
var _ = Describe("Upgrade testing", Label("nfr", "upgrade"), func() {
	var (
		files = []string{
			"ngf-upgrade/cafe.yaml",
			"ngf-upgrade/cafe-secret.yaml",
			"ngf-upgrade/gateway.yaml",
			"ngf-upgrade/cafe-routes.yaml",
		}

		ns = &core.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "ngf-upgrade",
			},
		}

		valuesFile  = "manifests/ngf-upgrade/values.yaml"
		resultsFile *os.File
		resultsDir  string
	)

	BeforeEach(func() {
		// this test is unique in that it needs to install the previous version of NGF,
		// so we need to uninstall the version installed at the suite level, then install the custom version
		teardown(releaseName)

		cfg := setupConfig{
			chartPath:    "oci://ghcr.io/nginxinc/charts/nginx-gateway-fabric",
			gwAPIVersion: *gatewayAPIPrevVersion,
			deploy:       true,
		}
		setup(cfg, "--values", valuesFile)

		Expect(resourceManager.Apply([]client.Object{ns})).To(Succeed())
		Expect(resourceManager.ApplyFromFiles(files, ns.Name)).To(Succeed())
		Expect(resourceManager.WaitForAppsToBeReady(ns.Name)).To(Succeed())

		var err error
		resultsDir, err = framework.CreateResultsDir("ngf-upgrade", version)
		Expect(err).ToNot(HaveOccurred())

		filename := filepath.Join(resultsDir, fmt.Sprintf("%s.md", version))
		resultsFile, err = framework.CreateResultsFile(filename)
		Expect(err).ToNot(HaveOccurred())
		Expect(framework.WriteSystemInfoToFile(resultsFile, clusterInfo, *plusEnabled)).To(Succeed())
	})

	AfterEach(func() {
		Expect(resourceManager.DeleteFromFiles(files, ns.Name)).To(Succeed())
		Expect(resourceManager.Delete([]client.Object{ns})).To(Succeed())
		resultsFile.Close()
	})

	It("upgrades NGF with zero downtime", func() {
		cfg := framework.InstallationConfig{
			ReleaseName:          releaseName,
			Namespace:            ngfNamespace,
			ChartPath:            localChartPath,
			NgfImageRepository:   *ngfImageRepository,
			NginxImageRepository: *nginxImageRepository,
			ImageTag:             *imageTag,
			ImagePullPolicy:      *imagePullPolicy,
			ServiceType:          *serviceType,
			IsGKEInternalLB:      *isGKEInternalLB,
		}

		type metricsResults struct {
			metrics  *framework.Metrics
			testName string
			scheme   string
		}
		metricsCh := make(chan *metricsResults, 2)
		var wg sync.WaitGroup

		type testCfg struct {
			desc   string
			port   string
			target framework.Target
		}

		tests := []testCfg{
			{
				desc: "Send http /coffee traffic",
				port: "80",
				target: framework.Target{
					Method: "GET",
					URL:    "http://cafe.example.com/coffee",
				},
			},
			{
				desc: "Send https /tea traffic",
				port: "443",
				target: framework.Target{
					Method: "GET",
					URL:    "https://cafe.example.com/tea",
				},
			},
		}

		for _, test := range tests {
			wg.Add(1)
			go func(cfg testCfg) {
				defer GinkgoRecover()
				defer wg.Done()

				loadTestCfg := framework.LoadTestConfig{
					Targets:     []framework.Target{cfg.target},
					Rate:        100,
					Duration:    60 * time.Second,
					Description: cfg.desc,
					Proxy:       fmt.Sprintf("%s:%s", address, cfg.port),
					ServerName:  "cafe.example.com",
				}

				results, metrics := framework.RunLoadTest(loadTestCfg)

				scheme := strings.Split(cfg.target.URL, "://")[0]
				metricsRes := metricsResults{
					metrics:  &metrics,
					testName: fmt.Sprintf("\n## Test: %s\n\n```text\n", cfg.desc),
					scheme:   scheme,
				}

				buf := new(bytes.Buffer)
				encoder := framework.NewCSVEncoder(buf)
				for _, res := range results {
					res := res
					Expect(encoder.Encode(&res)).To(Succeed())
				}

				csvName := fmt.Sprintf("%s.csv", scheme)
				filename := filepath.Join(resultsDir, csvName)
				csvFile, err := framework.CreateResultsFile(filename)
				Expect(err).ToNot(HaveOccurred())

				_, err = fmt.Fprint(csvFile, buf.String())
				Expect(err).ToNot(HaveOccurred())
				csvFile.Close()

				output, err := framework.GeneratePNG(resultsDir, csvName, fmt.Sprintf("%s.png", scheme))
				Expect(err).ToNot(HaveOccurred(), string(output))

				metricsCh <- &metricsRes
			}(test)
		}

		// allow traffic flow to start
		time.Sleep(2 * time.Second)

		// update Gateway API and NGF
		output, err := framework.InstallGatewayAPI(k8sClient, *gatewayAPIVersion, *k8sVersion)
		Expect(err).ToNot(HaveOccurred(), string(output))

		output, err = framework.UpgradeNGF(cfg, "--values", valuesFile)
		Expect(err).ToNot(HaveOccurred(), string(output))

		Expect(resourceManager.ApplyFromFiles([]string{"ngf-upgrade/gateway-updated.yaml"}, ns.Name)).To(Succeed())

		podNames, err := framework.GetReadyNGFPodNames(k8sClient, ngfNamespace, releaseName, timeoutConfig.GetTimeout)
		Expect(err).ToNot(HaveOccurred())
		Expect(podNames).ToNot(BeEmpty())

		// ensure that the leader election lease has been updated to the new pods
		leaseCtx, leaseCancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer leaseCancel()

		var lease coordination.Lease
		key := types.NamespacedName{Name: "ngf-test-nginx-gateway-fabric-leader-election", Namespace: ngfNamespace}
		Expect(wait.PollUntilContextCancel(
			leaseCtx,
			500*time.Millisecond,
			true, /* poll immediately */
			func(_ context.Context) (bool, error) {
				Expect(k8sClient.Get(leaseCtx, key, &lease)).To(Succeed())

				if lease.Spec.HolderIdentity != nil {
					for _, podName := range podNames {
						if podName == *lease.Spec.HolderIdentity {
							return true, nil
						}
					}
				}

				return false, nil
			},
		)).To(Succeed())

		// ensure that the Gateway has been properly updated with a new listener
		gwCtx, gwCancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer gwCancel()

		var gw v1.Gateway
		key = types.NamespacedName{Name: "gateway", Namespace: ns.Name}
		Expect(wait.PollUntilContextCancel(
			gwCtx,
			500*time.Millisecond,
			true, /* poll immediately */
			func(ctx context.Context) (bool, error) {
				Expect(k8sClient.Get(ctx, key, &gw)).To(Succeed())
				expListenerName := "http-new"
				for _, listener := range gw.Status.Listeners {
					if listener.Name == v1.SectionName(expListenerName) {
						return true, nil
					}
				}
				return false, nil
			},
		)).To(Succeed())

		wg.Wait()
		close(metricsCh)

		// write out the results
		for res := range metricsCh {
			_, err := fmt.Fprint(resultsFile, res.testName)
			Expect(err).ToNot(HaveOccurred())

			Expect(framework.WriteResults(resultsFile, res.metrics)).To(Succeed())

			_, err = fmt.Fprintf(resultsFile, "```\n\n![%[1]v.png](%[1]v.png)\n", res.scheme)
			Expect(err).ToNot(HaveOccurred())
		}
	})
})
