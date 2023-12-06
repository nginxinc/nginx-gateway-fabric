#!/bin/bash

source utils/vars.env

gcloud compute instances delete ${VM_NAME} --project=${GKE_PROJECT} --zone=${GKE_CLUSTER_ZONE}
