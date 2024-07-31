#!/usr/bin/env bash

num_namespaces=$1

# Create namespaces
for ((i=1; i<=$num_namespaces; i++)); do
    namespace_name="namespace$i"
    kubectl create namespace "$namespace_name"
done

# Create single instance resources
kubectl create -f certificate-ns-and-cafe-secret.yaml
kubectl create -f reference-grant.yaml

# Create backend service and apps
for ((i=1; i<=$num_namespaces; i++)); do
    namespace_name="namespace$i"
    sed -e "s/coffee/coffee${namespace_name}/g" -e "s/tea/tea${namespace_name}/g" cafe.yaml | kubectl apply -n "$namespace_name" -f -
done

# Create routes
for ((i=1; i<=$num_namespaces; i++)); do
    namespace_name="namespace$i"
    sed -e "s/coffee/coffee${namespace_name}/g" -e "s/tea/tea${namespace_name}/g" cafe-routes.yaml | kubectl apply -n "$namespace_name" -f -
done

# Wait for apps to be ready
sleep 60

# Create Gateway
kubectl create -f gateway.yaml
