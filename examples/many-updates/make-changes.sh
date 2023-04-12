#!/bin/bash

kubectl apply -f gateway.yaml
kubectl apply -f cafe.yaml

trap "echo 'Ctrl+C pressed. Exiting...'; exit" INT

while true
do
    echo "Scaling up..."
    for i in {1..6}; do
      kubectl scale --replicas ${i} deployment tea
      kubectl apply -f "${i}-cafe-routes.yaml"
      kubectl scale --replicas ${i} deployment coffee
    done

    sleep 1
    echo "Scaling down..."
    for i in {6..1}; do
      kubectl scale --replicas ${i} deployment tea
      kubectl apply -f "${i}-cafe-routes.yaml"
      kubectl scale --replicas ${i} deployment coffee
    done
done
