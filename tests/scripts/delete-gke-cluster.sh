#!/usr/bin/env bash

source scripts/vars.env

gcloud container clusters delete ${GKE_CLUSTER_NAME} --zone ${GKE_CLUSTER_ZONE} --project ${GKE_PROJECT} --quiet
