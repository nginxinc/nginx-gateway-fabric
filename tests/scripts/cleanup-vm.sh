#!/bin/bash

source scripts/vars.env
gcloud compute instances delete ${VM_NAME} --project=${GKE_PROJECT} --zone=${GKE_CLUSTER_ZONE}
gcloud compute firewall-rules delete ${FIREWALL_RULE_NAME} --project=${GKE_PROJECT}
