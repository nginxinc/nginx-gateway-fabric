#!/bin/bash

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source utils/vars.env

gcloud compute instances create ${VM_NAME} --project=${GKE_PROJECT} --zone=${GKE_CLUSTER_ZONE} --machine-type=e2-medium \
    --network-interface=network-tier=PREMIUM,stack-type=IPV4_ONLY,subnet=default --maintenance-policy=MIGRATE \
    --provisioning-model=STANDARD --service-account=${GKE_SVC_ACCOUNT} \
    --scopes=https://www.googleapis.com/auth/devstorage.read_only,https://www.googleapis.com/auth/logging.write,https://www.googleapis.com/auth/monitoring.write,https://www.googleapis.com/auth/servicecontrol,https://www.googleapis.com/auth/service.management.readonly,https://www.googleapis.com/auth/trace.append,https://www.googleapis.com/auth/cloud-platform \
    --tags=${TAGS} --create-disk=auto-delete=yes,boot=yes,device-name=${VM_NAME},image=${IMAGE},mode=rw,size=10 --no-shielded-secure-boot --shielded-vtpm --shielded-integrity-monitoring --labels=goog-ec-src=vm_add-gcloud --reservation-affinity=any

# Give the VM a chance to come up before attempting ssh connection
sleep 15

gcloud compute scp --zone ${GKE_CLUSTER_ZONE} ${SCRIPT_DIR}/vars.env ${VM_NAME}:~

gcloud compute ssh --zone ${GKE_CLUSTER_ZONE} ${VM_NAME} --command="bash -s" < ${SCRIPT_DIR}/remote-scripts/install-deps-and-run-tests.sh

gcloud compute scp --zone ${GKE_CLUSTER_ZONE} --recurse ${VM_NAME}:~/nginx-gateway-fabric/tests/suite/results .

cp -r results/* suite/results && rm -rf results/
