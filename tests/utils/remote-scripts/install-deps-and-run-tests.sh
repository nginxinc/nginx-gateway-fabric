#!/bin/bash

set -e

source ~/vars.env

sudo apt-get -y update && sudo apt-get -y install git make kubectl google-cloud-sdk-gke-gcloud-auth-plugin wrk jq && \
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash && \
export GO_VERSION=$(curl -sSL "https://golang.org/dl/?mode=json" | jq -r '.[0].version') && \
wget https://go.dev/dl/${GO_VERSION}.linux-amd64.tar.gz && \
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf ${GO_VERSION}.linux-amd64.tar.gz && \
echo "export PATH=$PATH:/usr/local/go/bin" >> $HOME/.profile && . $HOME/.profile && \
rm -rf ${GO_VERSION}.linux-amd64.tar.gz && \
sudo chmod 666 /etc/hosts && \
gcloud container clusters get-credentials ${GKE_CLUSTER_NAME} --zone ${GKE_CLUSTER_ZONE}

git clone https://github.com/ciarams87/nginx-gateway-fabric.git && cd nginx-gateway-fabric && git checkout tests/automate-dp-perf && cd tests

make test TAG=${TAG} PREFIX=${PREFIX} PULL_POLICY=Always GW_SERVICE_TYPE=LoadBalancer GW_SVC_GKE_INTERNAL=true
