VERSION = edge
TAG = $(VERSION)
PREFIX ?= nginx-kubernetes-gateway

GIT_COMMIT = $(shell git rev-parse HEAD)
DATE = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

TARGET ?= local

KIND_KUBE_CONFIG_FOLDER = $${HOME}/.kube/kind
OUT_DIR=$(shell pwd)/build/.out

.DEFAULT_GOAL := help

AGENT_VERSION ?= 2.22.1
ALPINE_VERSION ?= 3.16
NGINX_WITH_AGENT_PREFIX ?= nginx-with-agent

.PHONY: help
help: Makefile ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "; printf "Usage:\n\n    make \033[36m<target>\033[0m\n\nTargets:\n\n"}; {printf "    \033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: container
container: build ## Build the container
	@docker -v || (code=$$?; printf "\033[0;31mError\033[0m: there was a problem with Docker\n"; exit $$code)
	docker build --build-arg VERSION=$(VERSION) --build-arg GIT_COMMIT=$(GIT_COMMIT) --build-arg DATE=$(DATE) --target $(TARGET) -f build/Dockerfile -t $(PREFIX):$(TAG) .

.PHONY: nginx-with-agent-container
nginx-with-agent-container: ## Build the nginx-with-agent container
	@docker -v || (code=$$?; printf "\033[0;31mError\033[0m: there was a problem with Docker\n"; exit $$code)
	docker build --build-arg AGENT_VERSION=$(AGENT_VERSION) --build-arg ALPINE_VERSION=$(ALPINE_VERSION) -f build/nginx-with-agent/Dockerfile -t $(PREFIX)/$(NGINX_WITH_AGENT_PREFIX):$(TAG) .

.PHONY: build
build: ## Build the binary
ifeq (${TARGET},local)
	@go version || (code=$$?; printf "\033[0;31mError\033[0m: unable to build locally\n"; exit $$code)
	CGO_ENABLED=0 GOOS=linux go build -trimpath -a -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${GIT_COMMIT} -X main.date=${DATE}" -o $(OUT_DIR)/gateway github.com/nginxinc/nginx-kubernetes-gateway/cmd/gateway
endif

.PHONY: generate
generate: ## Run go generate
	go generate ./...

.PHONY: clean
clean: ## Clean the build
	-rm -r $(OUT_DIR)

.PHONY: clean--go-cache
clean-go-cache: ## Clean go cache
	@go clean -modcache

.PHONY: deps
deps: ## Add missing and remove unused modules, verify deps and download them to local cache
	@go mod tidy && go mod verify && go mod download

.PHONY: create-kind-cluster
create-kind-cluster: ## Create a kind cluster
	kind create cluster --image kindest/node:v1.27.1
	kind export kubeconfig --kubeconfig $(KIND_KUBE_CONFIG_FOLDER)/config

.PHONY: delete-kind-cluster
delete-kind-cluster: ## Delete kind cluster
	kind delete cluster

.PHONY: fmt
fmt: ## Run go fmt against code
	go fmt ./...

.PHONY: njs-fmt
njs-fmt: ## Run prettier against the njs httpmatches module
	docker run --rm -w /modules \
		-v $(PWD)/internal/nginx/modules/:/modules/ \
		node:18 \
		/bin/bash -c "npm install && npm run format"

.PHONY: vet
vet: ## Run go vet against code
	go vet ./...

.PHONY: lint
lint: ## Run golangci-lint against code
	docker run --pull always --rm -v $(shell pwd):/nginx-kubernetes-gateway -w /nginx-kubernetes-gateway -v $(shell go env GOCACHE):/cache/go -e GOCACHE=/cache/go -e GOLANGCI_LINT_CACHE=/cache/go -v $(shell go env GOPATH)/pkg:/go/pkg golangci/golangci-lint:latest golangci-lint --color always run

.PHONY: unit-test
unit-test: ## Run unit tests for the go code
	go test ./... -race -coverprofile cover.out
	go tool cover -html=cover.out -o cover.html

njs-unit-test: ## Run unit tests for the njs httpmatches module
	docker run --rm -w /modules \
		-v $(PWD)/internal/nginx/modules:/modules/ \
		node:18 \
		/bin/bash -c "npm install && npm test && npm run clean"

.PHONY: dev-all
dev-all: deps fmt njs-fmt vet lint unit-test njs-unit-test ## Run all the development checks
