#!/bin/bash

set -e

source ~/vars.env

sudo apt-get -y update && sudo apt-get -y install git make kubectl google-cloud-sdk-gke-gcloud-auth-plugin jq gnuplot && \
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash && \
export GO_VERSION=$(curl -sSL "https://golang.org/dl/?mode=json" | jq -r '.[0].version') && \
wget https://go.dev/dl/${GO_VERSION}.linux-amd64.tar.gz && \
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf ${GO_VERSION}.linux-amd64.tar.gz && \
rm -rf ${GO_VERSION}.linux-amd64.tar.gz && \
gcloud container clusters get-credentials ${GKE_CLUSTER_NAME} --zone ${GKE_CLUSTER_ZONE} --project=${GKE_PROJECT}

git clone https://github.com/${NGF_REPO}/nginx-gateway-fabric.git && cd nginx-gateway-fabric/tests && git checkout ${NGF_BRANCH}
