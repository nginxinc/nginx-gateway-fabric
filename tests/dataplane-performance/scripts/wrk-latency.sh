#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

echo -e "# Results\n"

echo -e "## Test1: Running coffee path based routing\n"
wrk -t2 -c10 -d30 http://cafe.example.com/coffee --latency
echo -e "\n## Test2: Running coffee header based routing\n"
wrk -t2 -c10 -d30 http://cafe.example.com/coffee -H "version:v2" --latency
echo -e "\n## Test3: Running coffee query based routing\n"
wrk -t2 -c10 -d30 http://cafe.example.com/coffee?TEST=v2 --latency
echo -e "\n## Test4: Running tea GET method based routing\n"
wrk -t2 -c10 -d30 http://cafe.example.com/tea --latency
echo -e "\n## Test5: Running tea POST method based routing\n"
wrk -t2 -c10 -d30 http://cafe.example.com/tea -s ${SCRIPT_DIR}/post.lua --latency
