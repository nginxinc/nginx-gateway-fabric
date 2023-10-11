#!/bin/bash

num_namespaces=$1

# Delete namespaces
for ((i=1; i<=$num_namespaces; i++)); do
    namespace_name="namespace$i"
    kubectl delete namespace "$namespace_name"
done

# Delete single instance resources
kubectl delete -f gateway.yaml
kubectl delete -f reference-grant.yaml
kubectl delete -f certificate-ns-and-cafe-secret.yaml
