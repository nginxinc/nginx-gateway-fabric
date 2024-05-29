# syntax=docker/dockerfile:1.7
FROM golang:1.22

WORKDIR /go/src/github.com/nginxinc/nginx-gateway-fabric/tests/conformance

COPY ../. /go/src/github.com/nginxinc/nginx-gateway-fabric/
COPY --link go.mod /go/src/github.com/nginxinc/nginx-gateway-fabric/tests/
COPY --link go.sum /go/src/github.com/nginxinc/nginx-gateway-fabric/tests/
RUN go mod download

COPY --link conformance /go/src/github.com/nginxinc/nginx-gateway-fabric/tests/conformance
