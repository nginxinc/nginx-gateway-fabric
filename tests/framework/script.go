package framework

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func RunScript(path string, args []string) ([]byte, error) {
	args = append([]string{path}, args...)
	return exec.Command("/bin/bash", args...).Output()
}

func RunScriptOutputToFile(scriptPath string, args []string, outputPath string, ci ClusterInfo) error {
	output, err := RunScript(scriptPath, args)
	if err != nil {
		return fmt.Errorf("Received error running script %v: %w", scriptPath, err)
	}

	//nolint:gosec
	resFile, err := os.OpenFile(outputPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o777)
	if err != nil {
		return err
	}
	defer resFile.Close()

	err = writeSystemInfoToFile(resFile, ci)
	if err != nil {
		return fmt.Errorf("Could not write system info to file %v: %w", outputPath, err)
	}

	if _, err := resFile.Write(output); err != nil {
		return fmt.Errorf("Could not write script output to file %v: %w", outputPath, err)
	}
	return nil
}

func AddEntryToHostsFile(domain string, ip string) error {
	hostsFile, err := os.OpenFile("/etc/hosts", os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}

	defer hostsFile.Close()

	entry := fmt.Sprintf("%v %v\n", ip, domain)

	scanner := bufio.NewScanner(hostsFile)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, entry) {
			// Entry already exists in hosts file
			return nil
		}
	}

	if _, err := hostsFile.Write([]byte(entry)); err != nil {
		return err
	}
	return nil
}

func RemoveEntryFromHostsFile(domain string, ip string) error {
	entry := fmt.Sprintf("%v %v", ip, domain)
	file, err := os.Open("/etc/hosts")
	if err != nil {
		return err
	}
	defer file.Close()

	tempFile, err := os.CreateTemp("", "tempfile")
	if err != nil {
		return err
	}
	defer tempFile.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		if !strings.Contains(line, entry) {
			_, err := tempFile.WriteString(line + "\n")
			if err != nil {
				return err
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	err = os.Rename(tempFile.Name(), "/etc/hosts")
	if err != nil {
		return err
	}
	err = os.Chmod("/etc/hosts", 0o644)
	if err != nil {
		return err
	}

	return nil
}

func writeSystemInfoToFile(file *os.File, ci ClusterInfo) error {
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
