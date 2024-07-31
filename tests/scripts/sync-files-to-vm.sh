#!/usr/bin/env bash

source scripts/vars.env

NGF_DIR=$(dirname "$CUR")

gcloud compute config-ssh --ssh-config-file ngf-gcp.ssh > /dev/null

rsync -ave 'ssh -F ngf-gcp.ssh' ${NGF_DIR} username@${RESOURCE_NAME}.${GKE_CLUSTER_ZONE}.${GKE_PROJECT}:~
