#!/usr/bin/env bash

set -eo pipefail

CODE=$(kubectl get pod conformance -o jsonpath='{.status.containerStatuses[].state.terminated.exitCode}')
if [ "${CODE}" -ne 0 ]; then
    exit 2
fi
