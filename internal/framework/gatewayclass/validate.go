package gatewayclass

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/conditions"
)

const (
	BundleVersionAnnotation = "gateway.networking.k8s.io/bundle-version"
	RecommendedVersion      = "v1.0.0"
)

var gatewayCRDs = map[string]apiVersion{
	"gatewayclasses.gateway.networking.k8s.io":  {},
	"gateways.gateway.networking.k8s.io":        {},
	"httproutes.gateway.networking.k8s.io":      {},
	"referencegrants.gateway.networking.k8s.io": {},
}

type apiVersion struct {
	major string
	minor string
}

func ValidateCRDVersions(
	crdMetadata map[types.NamespacedName]*metav1.PartialObjectMetadata,
) ([]conditions.Condition, bool) {
	installedAPIVersions := getBundleVersions(crdMetadata)
	supportedAPIVersion := parseVersionString(RecommendedVersion)

	var unsupported, bestEffort bool

	for _, version := range installedAPIVersions {
		if version.major != supportedAPIVersion.major {
			unsupported = true
		} else if version.minor != supportedAPIVersion.minor {
			bestEffort = true
		}
	}

	if unsupported {
		return conditions.NewGatewayClassUnsupportedVersion(RecommendedVersion), false
	}

	if bestEffort {
		return conditions.NewGatewayClassSupportedVersionBestEffort(RecommendedVersion), true
	}

	return nil, true
}

func parseVersionString(version string) apiVersion {
	versionBits := strings.Split(version, ".")
	if len(versionBits) != 3 {
		return apiVersion{}
	}

	major, _ := strings.CutPrefix(versionBits[0], "v")

	return apiVersion{
		major: major,
		minor: versionBits[1],
	}
}

func getBundleVersions(crdMetadata map[types.NamespacedName]*metav1.PartialObjectMetadata) map[string]apiVersion {
	versions := make(map[string]apiVersion)

	for nsname, md := range crdMetadata {
		if _, ok := gatewayCRDs[nsname.Name]; ok {
			bundleVersion := md.Annotations[BundleVersionAnnotation]
			versions[bundleVersion] = parseVersionString(bundleVersion)
		}
	}

	return versions
}
