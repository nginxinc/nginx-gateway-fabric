#!/bin/sh

set -e
set -x
set -euxo pipefail

handle_term()
{
    echo "received TERM signal"
    echo "stopping nginx-agent ..."
    kill -TERM "${agent_pid}" 2>/dev/null
    echo "stopping nginx ..."
    kill -TERM "${nginx_pid}" 2>/dev/null
}

trap 'handle_term' TERM

# Launch nginx
echo "starting nginx ..."
nginx -g "daemon off;" &

nginx_pid=$!

cat /etc/nginx-agent/nginx-agent.conf
# start nginx-agent, pass args
echo "starting nginx-agent ..."
nginx-agent "$@" &

agent_pid=$!

if [ $? != 0 ]; then
    echo "couldn't start the agent, please check the log file"
    exit 1
fi

wait_term()
{
    wait ${agent_pid}
    trap - TERM
    kill -QUIT "${nginx_pid}" 2>/dev/null
    echo "waiting for nginx to stop..."
    wait ${nginx_pid}
}

wait_term

echo "nginx-agent process has stopped, exiting."
