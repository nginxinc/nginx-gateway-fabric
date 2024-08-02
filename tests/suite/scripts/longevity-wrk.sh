#!/usr/bin/env bash

SVC_IP=$(kubectl -n nginx-gateway get svc ngf-longevity-nginx-gateway-fabric -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

echo "${SVC_IP} cafe.example.com" | sudo tee -a /etc/hosts

nohup wrk -t2 -c100 -d96h http://cafe.example.com/coffee &>~/coffee.txt &

nohup wrk -t2 -c100 -d96h https://cafe.example.com/tea &>~/tea.txt &
