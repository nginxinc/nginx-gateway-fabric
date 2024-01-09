#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source scripts/vars.env

gcloud compute scp --zone ${GKE_CLUSTER_ZONE} --project=${GKE_PROJECT} ${SCRIPT_DIR}/vars.env ${RESOURCE_NAME}:~

gcloud compute ssh --zone ${GKE_CLUSTER_ZONE} --project=${GKE_PROJECT} ${RESOURCE_NAME} --command="bash -s" < ${SCRIPT_DIR}/remote-scripts/run-tests.sh

gcloud compute scp --zone ${GKE_CLUSTER_ZONE} --project=${GKE_PROJECT} --recurse ${RESOURCE_NAME}:~/nginx-gateway-fabric/tests/results .
