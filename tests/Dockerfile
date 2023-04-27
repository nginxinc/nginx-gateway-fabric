FROM golang:1.20

WORKDIR /go/src/github.com/nginxinc/nginx-kubernetes-gateway/tests

COPY go.mod go.sum /go/src/github.com/nginxinc/nginx-kubernetes-gateway
RUN go mod download

COPY tests /go/src/github.com/nginxinc/nginx-kubernetes-gateway/tests