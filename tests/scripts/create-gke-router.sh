#!/bin/bash

source scripts/vars.env

gcloud compute routers create ${GKE_ROUTER_NAME} \
    --region ${GKE_CLUSTER_REGION} \
    --network default

gcloud compute routers nats create ${GKE_NATS_CONFIG_NAME} \
    --router-region ${GKE_CLUSTER_REGION} \
    --router ${GKE_ROUTER_NAME} \
    --nat-all-subnet-ip-ranges \
    --auto-allocate-nat-external-ips
