#!/usr/bin/env bash

source scripts/vars.env

CURRENT_AUTH_NETWORK=$(gcloud container clusters describe "${GKE_CLUSTER_NAME}" --zone="${GKE_CLUSTER_ZONE}" \
    --format="value(masterAuthorizedNetworksConfig.cidrBlocks[0])" | sed 's/cidrBlock=//')

gcloud container clusters update "${GKE_CLUSTER_NAME}" --zone="${GKE_CLUSTER_ZONE}" --enable-master-authorized-networks --master-authorized-networks="${SOURCE_IP_RANGE}","${CURRENT_AUTH_NETWORK}"
