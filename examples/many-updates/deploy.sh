#!/bin/bash

kubectl apply -f gateway.yaml
kubectl apply -f cafe.yaml

for i in {1..6}; do
  kubectl scale --replicas ${i} deployment tea
  kubectl apply -f "${i}-cafe-routes.yaml"
  kubectl scale --replicas ${i} deployment coffee
  sleep 0.1
done
