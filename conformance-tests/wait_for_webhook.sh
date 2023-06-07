#!/bin/sh

export exit=1
export count=0
while (( exit == 1 )) && (( count < 10)); do
    sleep 1
    count=$((count+1))
    echo waiting $count seconds for webhook container
    kubectl get events -n gateway-system | grep "Started container webhook"
    exit=`echo $?`
done