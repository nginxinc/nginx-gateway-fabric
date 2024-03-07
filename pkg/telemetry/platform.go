package telemetry

import (
	"strings"

	v1 "k8s.io/api/core/v1"
)

const (
	openshiftIdentifier = "node.openshift.io/os_id"
	k3sIdentifier       = "k3s"
	awsIdentifier       = "aws"
	gkeIdentifier       = "gce"
	azureIdentifier     = "azure"
	kindIdentifier      = "kind"
	rancherIdentifier   = "cattle-system"

	clusterPlatformGKE       = "gke"
	clusterPlatformAWS       = "eks"
	clusterPlatformAzure     = "aks"
	clusterPlatformKind      = "kind"
	clusterPlatformK3S       = "k3s"
	clusterPlatformOpenShift = "openshift"
	clusterPlatformRancher   = "rancher"
	clusterPlatformOther     = "other"
)

func CollectK8sPlatform(node v1.Node, namespaces v1.NamespaceList) string {
	if result := isMultiplePlatforms(node, namespaces); result != "" {
		return result
	}

	if isAWSPlatform(node) {
		return clusterPlatformAWS
	}
	if isGKEPlatform(node) {
		return clusterPlatformGKE
	}
	if isAzurePlatform(node) {
		return clusterPlatformAzure
	}
	if isKindPlatform(node) {
		return clusterPlatformKind
	}
	if isK3SPlatform(node) {
		return clusterPlatformK3S
	}

	return clusterPlatformOther
}

// isMultiplePlatforms checks for platforms that run on other platforms. e.g. Rancher on K3s.
func isMultiplePlatforms(node v1.Node, namespaces v1.NamespaceList) string {
	if isRancherPlatform(namespaces) {
		return clusterPlatformRancher
	}

	if isOpenshiftPlatform(node) {
		return clusterPlatformOpenShift
	}

	return ""
}

// For each of these, if we want to we can do both check the providerID AND check labels/annotations,
// I'm not too sure why we would want to do BOTH.
//
// I think doing both would add a greater certainty of a specific platform, however will potentially add to upkeep
// where if either the label/annotation or providerID changes it will mess this up and may group more clusters in
// the "Other" platform if they messed with any of the node labels/annotations.

func isOpenshiftPlatform(node v1.Node) bool {
	// openshift platform won't show up in node's ProviderID
	value, ok := node.Labels[openshiftIdentifier]

	return ok && value != ""
}

func isK3SPlatform(node v1.Node) bool {
	return strings.HasPrefix(node.Spec.ProviderID, k3sIdentifier)
}

func isAWSPlatform(node v1.Node) bool {
	return strings.HasPrefix(node.Spec.ProviderID, awsIdentifier)
}

func isGKEPlatform(node v1.Node) bool {
	return strings.HasPrefix(node.Spec.ProviderID, gkeIdentifier)
}

func isAzurePlatform(node v1.Node) bool {
	return strings.HasPrefix(node.Spec.ProviderID, azureIdentifier)
}

func isKindPlatform(node v1.Node) bool {
	return strings.HasPrefix(node.Spec.ProviderID, kindIdentifier)
}

func isRancherPlatform(namespaces v1.NamespaceList) bool {
	// rancher platform won't show up in the node's ProviderID
	for _, ns := range namespaces.Items {
		if ns.Name == rancherIdentifier {
			return true
		}
	}

	return false
}
