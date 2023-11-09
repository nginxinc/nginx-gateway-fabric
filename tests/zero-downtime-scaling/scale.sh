#!/bin/bash

scale_deployment() {
  replicas="$1"
  wait="$2"

  namespace="nginx-gateway"
  deployment_name=$(kubectl get deployment -n "$namespace" -o=jsonpath='{.items[0].metadata.name}')

  # Scale.
  echo "Scaling deployment to $replicas replicas..."
  kubectl scale deployment -n "$namespace" "$deployment_name" --replicas="$replicas"

 if [ "$wait" == "true" ]; then
    echo "Waiting for all Pods to be ready"
    kubectl wait pod --all --for=condition=Ready -n "$namespace"
    echo "All $replicas Pods are ready"
 else
    echo "Sleeping for 40 seconds to wait for Pod to terminate"
    sleep 40
 fi

  # Get the current observed generation of Gateway
  current_gen=$(kubectl get gateway gateway -o=jsonpath='{.status.conditions[0].observedGeneration}')
  next_gen=$((current_gen+1))

  # Apply the updated gateway.
  echo "Applying the updated gateway"
  kubectl apply -f "manifests/gateway-$replicas.yaml"

  for i in $(seq 1 60); do
    gen=$(kubectl get gateway gateway -o=jsonpath='{.status.conditions[0].observedGeneration}')
    if [[ $gen -ne $next_gen ]]; then
        echo "Observed generation is $gen; expected $next_gen; waiting"
        sleep 1
        continue
    else
        echo "Observed generation updated"
        break
    fi

    if [ $i -eq 60 ]; then
        echo "Observed generation not updated after 60 seconds. Exiting..."
        exit 1
    fi
  done

  kubectl get gateway gateway -o=jsonpath='{.status.conditions}'
}

if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <up|down>"
    exit 1
fi

if [ "$1" == "up" ]; then
    # Scale up
    for i in $(seq 2 25); do
        scale_deployment "$i" "true"
        sleep 1
    done
else
    # Scale down
    for i in $(seq 24 1); do
        scale_deployment "$i" "false"
        sleep 1
    done
fi
