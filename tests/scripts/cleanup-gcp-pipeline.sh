#!/bin/bash

source scripts/vars.env

gcloud compute instances delete ${RESOURCE_NAME} --quiet --project=${GKE_PROJECT} --zone=${GKE_CLUSTER_ZONE}
gcloud compute firewall-rules delete ${RESOURCE_NAME} --quiet --project=${GKE_PROJECT}

gcloud container clusters delete ${GKE_CLUSTER_NAME} --zone ${GKE_CLUSTER_ZONE} --project ${GKE_PROJECT} --quiet
