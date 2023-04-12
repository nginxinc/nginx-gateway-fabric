# syntax=docker/dockerfile:1.5
FROM golang:1.20 as builder
ARG VERSION
ARG GIT_COMMIT
ARG DATE

WORKDIR /go/src/github.com/nginxinc/nginx-kubernetes-gateway/cmd/gateway

COPY go.mod go.sum /go/src/github.com/nginxinc/nginx-kubernetes-gateway/
RUN go mod download

COPY cmd /go/src/github.com/nginxinc/nginx-kubernetes-gateway/cmd
COPY internal /go/src/github.com/nginxinc/nginx-kubernetes-gateway/internal
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -a -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${GIT_COMMIT} -X main.date=${DATE}" -o gateway .

FROM scratch as common
USER 1001:1001
ENTRYPOINT [ "/usr/bin/gateway" ]

FROM common as container
COPY --from=builder /go/src/github.com/nginxinc/nginx-kubernetes-gateway/cmd/gateway/gateway /usr/bin/

FROM common as local
COPY ./build/.out/gateway /usr/bin/

FROM common as goreleaser
COPY dist/gateway_linux_$TARGETARCH*/gateway /usr/bin/
