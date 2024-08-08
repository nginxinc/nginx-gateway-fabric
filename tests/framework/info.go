package framework

import (
	"fmt"

	core "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetLogs(rm ResourceManager, namespace string, releaseName string) string {
	var returnLogs string
	pods, err := rm.GetPods(namespace, client.MatchingLabels{
		"app.kubernetes.io/instance": releaseName,
	})
	if err != nil {
		return fmt.Sprintf("failed to get pods: %v", err)
	}
	for _, pod := range pods {
		for _, container := range pod.Spec.Containers {
			returnLogs += fmt.Sprintf("Logs for container %s:\n", container.Name)
			logs, err := rm.GetPodLogs(pod.Namespace, pod.Name, &core.PodLogOptions{
				Container: container.Name,
			})
			if err != nil {
				returnLogs += fmt.Sprintf("  failed to get logs: %v\n", err)
				continue
			}
			returnLogs += fmt.Sprintf("  %s\n", logs)
		}
	}
	return returnLogs
}

func GetEvents(rm ResourceManager, namespace string) string {
	var returnEvents string
	events, err := rm.GetEvents(namespace)
	if err != nil {
		return fmt.Sprintf("failed to get events: %v", err)
	}
	eventGroups := make(map[string][]core.Event)
	for _, event := range events.Items {
		eventGroups[event.InvolvedObject.Name] = append(eventGroups[event.InvolvedObject.Name], event)
	}
	for name, events := range eventGroups {
		returnEvents += fmt.Sprintf("Events for %s:\n", name)
		for _, event := range events {
			returnEvents += fmt.Sprintf("  %s\n", event.Message)
		}
		returnEvents += "\n"
	}
	return returnEvents
}
