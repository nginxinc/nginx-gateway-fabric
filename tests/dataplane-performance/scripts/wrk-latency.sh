#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

echo -e "# Results\n"

echo -e "## Test1: Running latte path based routing\n"
echo -e '```console'
wrk -t2 -c10 -d30 http://cafe.example.com/latte --latency
echo -e '```'
echo -e "\n## Test2: Running coffee header based routing\n"
echo -e '```console'
wrk -t2 -c10 -d30 http://cafe.example.com/coffee -H "version: v2" --latency
echo -e '```'
echo -e "\n## Test3: Running coffee query based routing\n"
echo -e '```console'
wrk -t2 -c10 -d30 http://cafe.example.com/coffee?TEST=v2 --latency
echo -e '```'
echo -e "\n## Test4: Running tea GET method based routing\n"
echo -e '```console'
wrk -t2 -c10 -d30 http://cafe.example.com/tea --latency
echo -e '```'
echo -e "\n## Test5: Running tea POST method based routing\n"
echo -e '```console'
wrk -t2 -c10 -d30 http://cafe.example.com/tea -s ${SCRIPT_DIR}/post.lua --latency
echo -e '```'
