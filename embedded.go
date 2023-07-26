package embeddedfiles

import _ "embed"

// StaticModeDeploymentYAML contains the YAML manifest of the Deployment resource for the static mode.
//
//go:embed conformance/provisioner/static-deployment.yaml
var StaticModeDeploymentYAML []byte
