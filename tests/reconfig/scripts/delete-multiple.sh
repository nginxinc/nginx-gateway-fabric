#!/usr/bin/env bash

num_namespaces=$1

# Delete namespaces
namespaces=""
for ((i = 1; i <= num_namespaces; i++)); do
    namespaces+="namespace${i} "
done

kubectl delete namespace "${namespaces}"

# Delete single instance resources
kubectl delete -f gateway.yaml
kubectl delete -f reference-grant.yaml
kubectl delete -f certificate-ns-and-cafe-secret.yaml
