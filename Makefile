# variables that should not be overridden by the user
VERSION = edge
GIT_COMMIT = $(shell git rev-parse HEAD || echo "unknown")
DATE = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
MANIFEST_DIR = $(shell pwd)/deploy/manifests 
NJS_DIR = $(shell pwd)/internal/mode/static/nginx/modules/src
CHART_DIR = $(shell pwd)/deploy/helm-chart

# variables that can be overridden by the user
PREFIX ?= nginx-kubernetes-gateway## The name of the image. For example, nginx-kubernetes-gateway
TAG ?= $(VERSION:v%=%)## The tag of the image. For example, 0.3.0
TARGET ?= local## The target of the build. Possible values: local and container
KIND_KUBE_CONFIG=$${HOME}/.kube/kind/config## The location of the kind kubeconfig
OUT_DIR ?= $(shell pwd)/build/out## The folder where the binary will be stored
ARCH ?= amd64## The architecture of the image and/or binary. For example: amd64 or arm64
override DOCKER_BUILD_OPTIONS += --build-arg VERSION=$(VERSION) --build-arg GIT_COMMIT=$(GIT_COMMIT) --build-arg DATE=$(DATE)## The options for the docker build command. For example, --pull

.DEFAULT_GOAL := help

.PHONY: help
help: Makefile ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "; printf "Usage:\n\n    make \033[36m<target>\033[0m [VARIABLE=value...]\n\nTargets:\n\n"}; {printf "    \033[36m%-30s\033[0m %s\n", $$1, $$2}'
	@grep -E '^(override )?[a-zA-Z_-]+ \??\+?= .*?## .*$$' $< | sort | awk 'BEGIN {FS = " \\??\\+?= .*?## "; printf "\nVariables:\n\n"}; {gsub(/override /, "", $$1); printf "    \033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: container
container: build ## Build the container
	@docker -v || (code=$$?; printf "\033[0;31mError\033[0m: there was a problem with Docker\n"; exit $$code)
	docker build --platform linux/$(ARCH) $(strip $(DOCKER_BUILD_OPTIONS)) --target $(strip $(TARGET)) -f build/Dockerfile -t $(strip $(PREFIX)):$(strip $(TAG)) .

.PHONY: build
build: ## Build the binary
ifeq (${TARGET},local)
	@go version || (code=$$?; printf "\033[0;31mError\033[0m: unable to build locally\n"; exit $$code)
	CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH) go build -trimpath -a -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${GIT_COMMIT} -X main.date=${DATE}" -o $(OUT_DIR)/gateway github.com/nginxinc/nginx-kubernetes-gateway/cmd/gateway
endif

.PHONY: build-goreleaser
build-goreleaser: ## Build the binary using GoReleaser
	@goreleaser -v || (code=$$?; printf "\033[0;31mError\033[0m: there was a problem with GoReleaser. Follow the docs to install it https://goreleaser.com/install\n"; exit $$code)
	GOOS=linux GOPATH=$(shell go env GOPATH) GOARCH=$(ARCH) goreleaser build --clean --snapshot --single-target

.PHONY: generate
generate: ## Run go generate
	go generate ./...

.PHONY: clean
clean: ## Clean the build
	-rm -r $(OUT_DIR)

.PHONY: clean-go-cache
clean-go-cache: ## Clean go cache
	@go clean -modcache

.PHONY: deps
deps: ## Add missing and remove unused modules, verify deps and download them to local cache
	@go mod tidy && go mod verify && go mod download

.PHONY: create-kind-cluster
create-kind-cluster: ## Create a kind cluster
	$(eval KIND_IMAGE=$(shell grep -m1 'FROM kindest/node' <conformance/tests/Dockerfile | awk -F'[ ]' '{print $$2}'))
	kind create cluster --image $(KIND_IMAGE)
	kind export kubeconfig --kubeconfig $(KIND_KUBE_CONFIG)

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
		-v $(PWD)/internal/mode/static/nginx/modules:/modules/ \
		node:18 \
		/bin/bash -c "npm install && npm test && npm run clean"

.PHONY: generate-njs-yaml
generate-njs-yaml: ## Generate the njs-modules ConfigMap
	kubectl create configmap njs-modules --from-file=$(NJS_DIR)/httpmatches.js --dry-run=client --output=yaml > $(strip $(MANIFEST_DIR))/njs-modules.yaml

.PHONY: fetch-crds-yaml
fetch-crds-yaml: ## Fetch the Gateway API resources yaml and output it to the Helm chart crds folder
	curl -s -L https://github.com/kubernetes-sigs/gateway-api/releases/download/v$(strip $(GW_API_VERSION))/standard-install.yaml -o $(CHART_DIR)/crds/gateway-crds.yaml

.PHONY: lint-helm
lint-helm: ## Run the helm chart linter
	helm lint $(CHART_DIR)

.PHONY: dev-all
dev-all: deps fmt njs-fmt vet lint unit-test njs-unit-test ## Run all the development checks
