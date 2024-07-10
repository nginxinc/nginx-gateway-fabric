package suite

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strconv"
	"text/template"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	v1 "sigs.k8s.io/gateway-api/apis/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/nginxinc/nginx-gateway-fabric/tests/framework"
)

var _ = Describe("Reconfiguration Performance Testing", Ordered, Label("reconfiguration"), func() {
	var ()

	BeforeAll(func() {
		resultsDir, err := framework.CreateResultsDir("reconfig", version)
		Expect(err).ToNot(HaveOccurred())

		filename := filepath.Join(resultsDir, framework.CreateResultsFilename("md", version, *plusEnabled))
		outFile, err := framework.CreateResultsFile(filename)
		Expect(err).ToNot(HaveOccurred())
		Expect(framework.WriteSystemInfoToFile(outFile, clusterInfo, *plusEnabled)).To(Succeed())

	})

	BeforeEach(func() {
		// possibly instead of teardown, can scale to 0 replicas.
		teardown(releaseName)

		setup(getDefaultSetupCfg())
	})

	AfterEach(func() {
		teardown(releaseName)
	})

	It("test 1", func() {
		Expect(createResourcesGWLast(30)).To(Succeed())
		Expect(checkResourceCreation(30)).To(Succeed())
		cleanupResources(30)
	})
	//
	//It("test 2", func() {
	//	Expect(createResourcesRoutesLast(30)).To(Succeed())
	//	Expect(checkResourceCreation(30)).To(Succeed())
	//	cleanupResources(30)
	//})
})

func createResourcesGWLast(resourceCount int) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutConfig.GetTimeout)
	defer cancel()

	for i := 1; i <= resourceCount; i++ {
		ns := core.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "namespace" + strconv.Itoa(i),
			},
		}
		Expect(k8sClient.Create(ctx, &ns)).To(Succeed())
	}

	ns := core.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "reconfig",
		},
	}
	Expect(resourceManager.Apply([]client.Object{&ns})).To(Succeed())
	Expect(resourceManager.ApplyFromFiles(
		[]string{
			"reconfig/certificate-ns-and-cafe-secret.yaml",
			"reconfig/reference-grant.yaml",
		},
		ns.Name)).To(Succeed())

	Expect(createUniqueResources(resourceCount, "manifests/reconfig/cafe.yaml")).To(Succeed())

	Expect(createUniqueResources(resourceCount, "manifests/reconfig/cafe-routes.yaml")).To(Succeed())

	time.Sleep(60 * time.Second)

	Expect(resourceManager.ApplyFromFiles([]string{"reconfig/gateway.yaml"}, ns.Name)).To(Succeed())

	return nil
}

func createResourcesRoutesLast(resourceCount int) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutConfig.GetTimeout)
	defer cancel()

	for i := 1; i <= resourceCount; i++ {
		ns := core.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "namespace" + strconv.Itoa(i),
			},
		}
		Expect(k8sClient.Create(ctx, &ns)).To(Succeed())
	}

	Expect(createUniqueResources(resourceCount, "manifests/reconfig/cafe.yaml")).To(Succeed())

	time.Sleep(60 * time.Second)

	ns := core.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "reconfig",
		},
	}
	Expect(resourceManager.Apply([]client.Object{&ns})).To(Succeed())
	Expect(resourceManager.ApplyFromFiles(
		[]string{
			"reconfig/certificate-ns-and-cafe-secret.yaml",
			"reconfig/reference-grant.yaml",
			"reconfig/gateway.yaml",
		},
		ns.Name)).To(Succeed())

	Expect(createUniqueResources(resourceCount, "manifests/reconfig/cafe-routes.yaml")).To(Succeed())

	return nil
}

func createUniqueResources(resourceCount int, fileName string) error {
	for i := 1; i <= resourceCount; i++ {
		nsName := "namespace" + strconv.Itoa(i)
		// Command to run sed and capture its output
		//nolint:gosec
		sedCmd := exec.Command("sed",
			"-e",
			"s/coffee/coffee"+nsName+"/g",
			"-e",
			"s/tea/tea"+nsName+"/g",
			fileName,
		)
		// Command to apply using kubectl
		kubectlCmd := exec.Command("kubectl", "apply", "-n", nsName, "-f", "-")

		sedOutput, err := sedCmd.Output()
		if err != nil {
			fmt.Println(err.Error() + ": " + string(sedOutput))
			return err
		}
		kubectlCmd.Stdin = bytes.NewReader(sedOutput)

		output, err := kubectlCmd.CombinedOutput()
		if err != nil {
			fmt.Println(err.Error() + ": " + string(output))
			return err
		}
	}
	return nil
}

func checkResourceCreation(resourceCount int) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutConfig.GetTimeout)
	defer cancel()

	var namespaces core.NamespaceList
	if err := k8sClient.List(ctx, &namespaces); err != nil {
		return fmt.Errorf("error getting namespaces: %w", err)
	}
	Expect(len(namespaces.Items)).To(BeNumerically(">=", resourceCount))

	var routes v1.HTTPRouteList
	if err := k8sClient.List(ctx, &routes); err != nil {
		return fmt.Errorf("error getting HTTPRoutes: %w", err)
	}
	Expect(len(routes.Items)).To(BeNumerically(">=", resourceCount*3))

	var pods core.PodList
	if err := k8sClient.List(ctx, &pods); err != nil {
		return fmt.Errorf("error getting Pods: %w", err)
	}
	Expect(len(pods.Items)).To(BeNumerically(">=", resourceCount*2))

	return nil
}

func cleanupResources(resourceCount int) {
	for i := 1; i <= resourceCount; i++ {
		nsName := "namespace" + strconv.Itoa(i)
		Expect(resourceManager.DeleteNamespace(nsName)).To(Succeed())
	}

	ns := core.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "reconfig",
		},
	}

	Expect(resourceManager.DeleteFromFiles([]string{
		"reconfig/certificate-ns-and-cafe-secret.yaml",
		"reconfig/reference-grant.yaml",
		"reconfig/gateway.yaml",
	}, ns.Name)).To(Succeed())

}

type reconfigTestResults struct {
	Name                 string
	EventsBuckets        []bucket
	ReloadBuckets        []bucket
	NumResources         int
	TimeToReadyTotal     int
	TimeToReadyAvgSingle int
	NGINXReloads         int
	NGINXReloadAvgTime   int
	EventsCount          int
	EventsAvgTime        int
}

const reconfigResultTemplate = `
## Test {{ .Name }}

- NumResources: {{ .NumResources }}

### Reloads and Time to Ready

- TimeToReadyTotal: {{ .TimeToReadyTotal }}
- TimeToReadyAvgSingle: {{ .TimeToReadyAvgSingle }}
- NGINX Reloads: {{ .NGINXReloads }}
- NGINX Reload Average Time: {{ .NGINXReloadAvgTime }}
- Reload distribution:
{{- range .ReloadBuckets }}
	- {{ .Le }}ms: {{ .Val }}
{{- end }}

### Event Batch Processing

- Event Batch Total: {{ .EventsCount }}
- Event Batch Processing Average Time: {{ .EventsAvgTime }}ms
- Event Batch Processing distribution:
{{- range .EventsBuckets }}
	- {{ .Le }}ms: {{ .Val }}
{{- end }}

`

func writeReconfigResults(dest io.Writer, results reconfigTestResults) error {
	tmpl, err := template.New("results").Parse(reconfigResultTemplate)
	if err != nil {
		return err
	}

	return tmpl.Execute(dest, results)
}
