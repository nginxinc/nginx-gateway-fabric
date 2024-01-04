#!/bin/bash

source scripts/vars.env

gcloud compute routers nats delete ${GKE_NATS_CONFIG_NAME} --router ${GKE_ROUTER_NAME} --router-region ${GKE_CLUSTER_REGION}
gcloud compute routers delete ${GKE_ROUTER_NAME} --region ${GKE_CLUSTER_REGION}
