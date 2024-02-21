# syntax=docker/dockerfile:1.6
FROM golang:1.22 as builder

WORKDIR /go/src/github.com/nginxinc/nginx-gateway-fabric

COPY go.mod go.sum /go/src/github.com/nginxinc/nginx-gateway-fabric
RUN go mod download

COPY . /go/src/github.com/nginxinc/nginx-gateway-fabric
RUN make build

FROM alpine:3.19 as capabilizer
RUN apk add --no-cache libcap

FROM capabilizer as local-capabilizer
COPY ./build/out/gateway /usr/bin/
RUN setcap 'cap_kill=+ep' /usr/bin/gateway

FROM capabilizer as container-capabilizer
COPY --from=builder /go/src/github.com/nginxinc/nginx-gateway-fabric/build/out/gateway /usr/bin/
RUN setcap 'cap_kill=+ep' /usr/bin/gateway

FROM capabilizer as goreleaser-capabilizer
ARG TARGETARCH
COPY dist/gateway_linux_$TARGETARCH*/gateway /usr/bin/
RUN setcap 'cap_kill=+ep' /usr/bin/gateway

FROM scratch as common
USER 102:1001
ARG BUILD_AGENT
ENV BUILD_AGENT=${BUILD_AGENT}
ENTRYPOINT [ "/usr/bin/gateway" ]

FROM common as container
COPY --from=container-capabilizer /usr/bin/gateway /usr/bin/

FROM common as local
COPY --from=local-capabilizer /usr/bin/gateway /usr/bin/

FROM common as goreleaser
COPY --from=goreleaser-capabilizer /usr/bin/gateway /usr/bin/
