FROM nginx:1.22.1-alpine
ARG AGENT_VERSION
ARG ALPINE_VERSION

WORKDIR /agent

RUN apk add --no-cache libcap

# For now, get the agent apk package from github release. Eventually, we will pull the pre-build package from nginx.org.
RUN wget -nv -O agent.apk https://github.com/nginx/agent/releases/download/v$AGENT_VERSION/nginx-agent-$AGENT_VERSION-v$ALPINE_VERSION-x86_64.apk \
    && apk add --allow-untrusted agent.apk

# Copy nginx-agent config file and entrypont script.
# We could also mount this to the Pod.
COPY ./build/agent/nginx-agent.conf /etc/nginx-agent/nginx-agent.conf
COPY ./build/agent/entrypoint.sh /agent/entrypoint.sh

# Copy nginx config file and httpmatches njs module.
# We could also mount this to the Pod.
COPY ./internal/nginx/modules/src/httpmatches.js /usr/lib/nginx/modules/njs/httpmatches.js
COPY ./build/agent/nginx.conf /etc/nginx/nginx.conf

# Create nginx directories.
RUN mkdir -p /etc/nginx/secrets /var/lib/nginx /var/log/nginx

# Clear /conf.d directory.
RUN rm -f /etc/nginx/conf.d/*

# Change owner of nginx directories to nginx user 101.
RUN chown -R 101:101 /etc/nginx /var/lib/nginx /var/log/nginx /var/cache/nginx

# Change owner of agent directories to nginx user 101.
RUN chown -R 101:101 /var/log/nginx-agent /etc/nginx-agent

# Make the entrypoint script executable.
RUN chmod +x /agent/entrypoint.sh

# The following instructions allow nginx and nginx-debug binaries to bind to privileged ports.
# However, adding this capability prevents the agent from reading nginx's /proc/<pid>/exe symlink which is required by
# the agent to determine the path to the nginx binary. While we wait for a more permanent fix, we will work around this
# by having nginx bind to non-privileged ports. See this write-up for more details: https://dxuuu.xyz/filecaps.html

#RUN setcap 'cap_net_bind_service=+ep' /usr/sbin/nginx 'cap_net_bind_service=+ep' /usr/sbin/nginx-debug
#RUN setcap -v 'cap_net_bind_service=+ep' /usr/sbin/nginx 'cap_net_bind_service=+ep' /usr/sbin/nginx-debug

# Set user to 101 (nginx)
USER 101:101

STOPSIGNAL SIGTERM

EXPOSE 8080 8443

ENTRYPOINT ["/agent/entrypoint.sh"]
