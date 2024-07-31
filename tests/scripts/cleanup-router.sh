#!/usr/bin/env bash

source scripts/vars.env

gcloud compute routers nats delete ${RESOURCE_NAME} --quiet --router ${RESOURCE_NAME} --router-region ${GKE_CLUSTER_REGION}
gcloud compute routers delete ${RESOURCE_NAME} --quiet --region ${GKE_CLUSTER_REGION}
