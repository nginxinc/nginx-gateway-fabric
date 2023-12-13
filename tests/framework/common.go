package framework

import (
	"fmt"
	"os"
)

// WriteSystemInfoToFile writes the cluster system info to the given file
func WriteSystemInfoToFile(file *os.File, ci ClusterInfo) error {
	clusterType := "Local"
	if ci.IsGKE {
		clusterType = "GKE"
	}
	text := fmt.Sprintf(
		//nolint:lll
		"# Results\n\n## Test environment\n\n%s Cluster:\n\n- Node count: %d\n- k8s version: %s\n- vCPUs per node: %d\n- RAM per node: %s\n- Max pods per node: %d\n",
		clusterType, ci.NodeCount, ci.K8sVersion, ci.CPUCountPerNode, ci.MemoryPerNode, ci.MaxPodsPerNode,
	)
	if _, err := fmt.Fprint(file, text); err != nil {
		return err
	}
	if ci.IsGKE {
		if _, err := fmt.Fprintf(file, "- Zone: %s\n- Instance Type: %s\n", ci.GkeZone, ci.GkeInstanceType); err != nil {
			return err
		}
	}
	return nil
}
