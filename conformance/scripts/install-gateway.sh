#!/bin/bash

if [ -z $1 ]; then
    echo "gateway API version argument not set; exiting"
    exit 1
fi

if [ -z $2 ]; then
    echo "install webhook argument not set; exiting"
    exit 1
fi

if [ $1 == "main" ]; then
    temp_dir=$(mktemp -d)
    cd ${temp_dir}
    curl -s https://codeload.github.com/kubernetes-sigs/gateway-api/tar.gz/main | tar -xz --strip=2 gateway-api-main/config
    kubectl apply -f crd/standard
    if [ $2 == "true" ]; then
        kubectl apply -f webhook
        kubectl wait --for=condition=available --timeout=60s deployment gateway-api-admission-server -n gateway-system
    fi
    rm -rf ${temp_dir}
else
    kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v$1/standard-install.yaml
    if [ $2 == "true" ]; then
        kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v$1/webhook-install.yaml
        kubectl wait --for=condition=available --timeout=60s deployment gateway-api-admission-server -n gateway-system
    fi
fi
