#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source scripts/vars.env

gcloud compute scp --zone ${GKE_CLUSTER_ZONE} --project=${GKE_PROJECT} ${SCRIPT_DIR}/vars.env username@${RESOURCE_NAME}:~

gcloud compute ssh --zone ${GKE_CLUSTER_ZONE} --project=${GKE_PROJECT} username@${RESOURCE_NAME} --command="bash -s" < ${SCRIPT_DIR}/remote-scripts/run-tests.sh

gcloud compute scp --zone ${GKE_CLUSTER_ZONE} --project=${GKE_PROJECT} --recurse username@${RESOURCE_NAME}:~/nginx-gateway-fabric/tests/results .
