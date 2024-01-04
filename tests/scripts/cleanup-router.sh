#!/bin/bash

source scripts/vars.env

gcloud compute routers nats delete ${RESOURCE_NAME} --router ${RESOURCE_NAME} --router-region ${GKE_CLUSTER_REGION}
gcloud compute routers delete ${RESOURCE_NAME} --region ${GKE_CLUSTER_REGION}
