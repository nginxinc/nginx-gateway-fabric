#!/bin/bash

source scripts/vars.env

ip_random_digit=$((1 + $RANDOM % 250))

gcloud container clusters create ${GKE_CLUSTER_NAME} \
    --project ${GKE_PROJECT} \
    --zone ${GKE_CLUSTER_ZONE} \
    --enable-master-authorized-networks \
    --enable-ip-alias \
    --service-account ${GKE_NODES_SERVICE_ACCOUNT} \
    --enable-private-nodes \
    --master-ipv4-cidr 172.16.${ip_random_digit}.32/28 \
    --metadata=block-project-ssh-keys=TRUE
