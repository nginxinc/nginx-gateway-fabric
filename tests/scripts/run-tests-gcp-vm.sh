#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source scripts/vars.env

gcloud compute ssh --zone ${GKE_CLUSTER_ZONE} ${VM_NAME} --command="bash -s" < ${SCRIPT_DIR}/remote-scripts/run-tests.sh

gcloud compute scp --zone ${GKE_CLUSTER_ZONE} --recurse ${VM_NAME}:~/nginx-gateway-fabric/tests/results .
