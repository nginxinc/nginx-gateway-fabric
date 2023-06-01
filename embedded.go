package embeddedfiles

import _ "embed"

// StaticModeDeploymentYAML contains the YAML manifest of the Deployment resource for the static mode.
//
// We put this in the root of the repo because goembed doesn't support relative/absolute paths and symlinks,
// and we want to keep the manifests in the deploy/manifests directory.
//
//go:embed deploy/manifests/deployment.yaml
var StaticModeDeploymentYAML []byte
