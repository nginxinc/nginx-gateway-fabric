//go:build !unit
// +build !unit

package tests

import (
	"flag"
	"path"
	"path/filepath"
	"runtime"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	k8sRuntime "k8s.io/apimachinery/pkg/runtime"
	ctlr "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/nginxinc/nginx-gateway-fabric/tests/framework"
)

func TestNGF(t *testing.T) {
	flag.Parse()
	if *gatewayAPIVersion == "" {
		panic("Gateway API version must be set")
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "NGF System Tests")
}

var k8sClient client.Client

var _ = BeforeSuite(func() {
	k8sConfig := ctlr.GetConfigOrDie()
	scheme := k8sRuntime.NewScheme()
	Expect(core.AddToScheme(scheme)).To(Succeed())
	Expect(apps.AddToScheme(scheme)).To(Succeed())
	Expect(apiext.AddToScheme(scheme)).To(Succeed())

	options := client.Options{
		Scheme: scheme,
	}

	var err error
	k8sClient, err = client.New(k8sConfig, options)
	Expect(err).ToNot(HaveOccurred())

	_, file, _, _ := runtime.Caller(0)
	fileDir := path.Join(path.Dir(file))
	basepath := filepath.Dir(fileDir)

	cfg := framework.InstallationConfig{
		ReleaseName:          "ngf-test",
		Namespace:            "nginx-gateway",
		ChartPath:            filepath.Join(basepath, "deploy/helm-chart"),
		NgfImageRepository:   *ngfImageRepository,
		NginxImageRepository: *nginxImageRepository,
		ImageTag:             *imageTag,
		ImagePullPolicy:      *imagePullPolicy,
		ServiceType:          "NodePort",
	}

	output, err := framework.InstallGatewayAPI(k8sClient, *gatewayAPIVersion, *k8sVersion)
	Expect(err).ToNot(HaveOccurred(), string(output))

	output, err = framework.InstallNGF(cfg)
	Expect(err).ToNot(HaveOccurred(), string(output))
})

var _ = AfterSuite(func() {
	cfg := framework.InstallationConfig{
		ReleaseName: "ngf-test",
		Namespace:   "nginx-gateway",
	}

	output, err := framework.UninstallNGF(cfg, k8sClient)
	Expect(err).ToNot(HaveOccurred(), string(output))

	output, err = framework.UninstallGatewayAPI(*gatewayAPIVersion, *k8sVersion)
	Expect(err).ToNot(HaveOccurred(), string(output))
})

var (
	gatewayAPIVersion = flag.String("gateway-api-version", "", "Version of Gateway API to install")
	k8sVersion        = flag.String("k8s-version", "latest", "Version of k8s being tested on")
	// Configurable NGF installation variables. Helm values will be used as defaults if not specified.
	ngfImageRepository   = flag.String("ngf-image-repo", "", "Image repo for NGF control plane")
	nginxImageRepository = flag.String("nginx-image-repo", "", "Image repo for NGF data plane")
	imageTag             = flag.String("image-tag", "", "Image tag for NGF images")
	imagePullPolicy      = flag.String("pull-policy", "", "Image pull policy for NGF images")
)
