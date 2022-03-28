FROM golang:1.18.0 as builder
ARG VERSION
ARG GIT_COMMIT
ARG DATE

WORKDIR /go/src/github.com/nginxinc/nginx-gateway-kubernetes/cmd/gateway

COPY go.mod go.sum /go/src/github.com/nginxinc/nginx-gateway-kubernetes
RUN go mod download

COPY cmd /go/src/github.com/nginxinc/nginx-gateway-kubernetes/cmd
COPY internal /go/src/github.com/nginxinc/nginx-gateway-kubernetes/internal
COPY pkg /go/src/github.com/nginxinc/nginx-gateway-kubernetes/pkg
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -a -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${GIT_COMMIT} -X main.date=${DATE}" -o ./build/.out/gateway .

FROM scratch as common
USER 1001:1001
ENTRYPOINT [ "/usr/bin/gateway" ]

FROM common as container
COPY --from=builder /go/src/github.com/nginxinc/nginx-gateway-kubernetes/cmd/gateway/gateway /usr/bin/

FROM common as local
COPY ./build/.out/gateway /usr/bin/
