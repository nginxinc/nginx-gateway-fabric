VERSION = 0.0.1
TAG = $(VERSION)
PREFIX ?= nginx-kubernetes-gateway

GIT_COMMIT = $(shell git rev-parse HEAD)
DATE = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

TARGET ?= local

KIND_KUBE_CONFIG_FOLDER = $${HOME}/.kube/kind
OUT_DIR=$(shell pwd)/build/.out

export DOCKER_BUILDKIT = 1

.PHONY: container
container: build
	docker build --build-arg VERSION=$(VERSION) --build-arg GIT_COMMIT=$(GIT_COMMIT) --build-arg DATE=$(DATE) --target $(TARGET) -f build/Dockerfile -t $(PREFIX):$(TAG) .

.PHONY: build
build:
ifeq (${TARGET},local)
	CGO_ENABLED=0 GOOS=linux go build -trimpath -a -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${GIT_COMMIT} -X main.date=${DATE}" -o $(OUT_DIR)/gateway github.com/nginxinc/nginx-kubernetes-gateway/cmd/gateway
endif

.PHONY: generate
generate:
	go generate ./...

.PHONY: out_dir
out_dir:
	mkdir -p $(OUT_DIR)

.PHONY: clean
clean: out_dir
	rm -rf $(OUT_DIR)

.PHONY: deps
deps:
	@go mod tidy && go mod verify && go mod download

.PHONY: update-codegen
update-codegen:
	# requires the root folder of the repo to be inside the GOPATH
	./hack/update-codegen.sh

.PHONY: verify-codegen
verify-codegen:
	# requires the root folder of the repo to be inside the GOPATH
	./hack/verify-codegen.sh

.PHONY: update-crds
update-crds: ## Update CRDs
	go run sigs.k8s.io/controller-tools/cmd/controller-gen crd:crdVersions=v1 schemapatch:manifests=./deploy/manifests/crds/ paths=./pkg/apis/... output:dir=./deploy/manifests/crds/

.PHONY: create-kind-cluster
create-kind-cluster:
	kind create cluster --image kindest/node:v1.22.1
	kind export kubeconfig --kubeconfig $(KIND_KUBE_CONFIG_FOLDER)/config

.PHONY: delete-kind-cluster
delete-kind-cluster:
	kind delete cluster

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: njs-fmt
njs-fmt: ## Run prettier against the njs httpmatches module.
	docker run --rm -w /modules \
		-v $(PWD)/internal/nginx/modules/:/modules/ \
		node:18 \
		/bin/bash -c "npm install && npm run format"

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: lint
lint: ## Run golangci-lint against code.
	docker run --pull always --rm -v $(shell pwd):/nginx-kubernetes-gateway -w /nginx-kubernetes-gateway -v $(shell go env GOCACHE):/cache/go -e GOCACHE=/cache/go -e GOLANGCI_LINT_CACHE=/cache/go -v $(shell go env GOPATH)/pkg:/go/pkg golangci/golangci-lint:latest golangci-lint --color always run

.PHONY: unit-test
unit-test: ## Run unit tests for the go code
	go test ./... -race -coverprofile cover.out
	go tool cover -html=cover.out -o cover.html

njs-unit-test: ## Run unit tests for the njs httpmatches module.
	docker run --rm -w /modules \
		-v $(PWD)/internal/nginx/modules:/modules/ \
		node:18 \
		/bin/bash -c "npm install && npm test && npm run clean"

.PHONY: dev-all
dev-all: deps fmt njs-fmt vet lint unit-test njs-unit-test
