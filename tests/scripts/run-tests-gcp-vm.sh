#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

NFR=${1:-false}

source scripts/vars.env

SCRIPT=run-tests.sh
if [ "${NFR}" = "true" ]; then
    SCRIPT=run-nfr-tests.sh
fi

gcloud compute scp --zone ${GKE_CLUSTER_ZONE} --project=${GKE_PROJECT} ${SCRIPT_DIR}/vars.env username@${RESOURCE_NAME}:~

gcloud compute ssh --zone ${GKE_CLUSTER_ZONE} --project=${GKE_PROJECT} username@${RESOURCE_NAME} \
    --command="export START_LONGEVITY=${START_LONGEVITY} &&\
        export STOP_LONGEVITY=${STOP_LONGEVITY} &&\
        bash -s" < ${SCRIPT_DIR}/remote-scripts/${SCRIPT}

if [ "${NFR}" = "true" ]; then
    gcloud compute scp --zone ${GKE_CLUSTER_ZONE} --project=${GKE_PROJECT} --recurse username@${RESOURCE_NAME}:~/nginx-gateway-fabric/tests/results .
fi
