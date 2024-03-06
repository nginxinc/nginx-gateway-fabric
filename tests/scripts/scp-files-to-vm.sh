#!/bin/bash

# Usage: bash /scp-files-to-vm.sh <path-to-local> <path-remote>
# e.g. bash scripts/scp-files-to-vm.sh framework/results.go /nginx-gateway-fabric/tests/framework/results.go
PATH_TO_LOCAL=$1

# The remote path will be appended to '~'. Requires leading /.
PATH_REMOTE=$2

source scripts/vars.env

gcloud compute scp --zone ${GKE_CLUSTER_ZONE} --project=${GKE_PROJECT} ${PATH_TO_LOCAL} username@${RESOURCE_NAME}:~${PATH_REMOTE}
