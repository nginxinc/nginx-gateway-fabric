package provisioner

import (
	"fmt"

	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// prepareDeployment prepares a new the static mode Deployment based on the YAML manifest.
// It will use the specified id to set unique parts of the deployment, so it must be unique among all Deployments for
// Gateways.
// It will configure the Deployment to use the Gateway with the given NamespacedName.
func prepareDeployment(depYAML []byte, id string, gwNsName types.NamespacedName) (*v1.Deployment, error) {
	dep := &v1.Deployment{}
	if err := yaml.Unmarshal(depYAML, dep); err != nil {
		return nil, fmt.Errorf("failed to unmarshal deployment: %w", err)
	}

	dep.ObjectMeta.Name = id
	dep.Spec.Selector.MatchLabels["app"] = id
	dep.Spec.Template.ObjectMeta.Labels["app"] = id

	extraArgs := []string{
		"--gateway=" + gwNsName.String(),
		"--update-gatewayclass-status=false",
	}
	dep.Spec.Template.Spec.Containers[0].Args = append(dep.Spec.Template.Spec.Containers[0].Args, extraArgs...)

	return dep, nil
}
