# syntax=docker/dockerfile:1.5

FROM golang:1.20

WORKDIR /go/src/github.com/nginxinc/nginx-kubernetes-gateway/conformance/tests/

COPY --link go.mod /go/src/github.com/nginxinc/nginx-kubernetes-gateway/
COPY --link go.sum /go/src/github.com/nginxinc/nginx-kubernetes-gateway/
RUN go mod download

COPY --link conformance/tests /go/src/github.com/nginxinc/nginx-kubernetes-gateway/conformance/tests/

CMD go test -timeout 25m -v . -tags conformance -args --gateway-class=${GATEWAY_CLASS} --debug --supported-features=${SUPPORTED_FEATURES} --exempt-features=${EXEMPT_FEATURES}
