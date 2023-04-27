KIND_KUBE_CONFIG_FOLDER = $${HOME}/.kube/kind
TAG = latest
PREFIX = test-runner

build:
	docker build -t $(PREFIX):$(TAG) -f Dockerfile ..

create-kind-cluster:
	kind create cluster --image kindest/node:v1.27.1
	kind export kubeconfig --kubeconfig $(KIND_KUBE_CONFIG_FOLDER)/config

delete-kind-cluster:
	kind delete cluster

load-image-to-kind:
	kind load docker-image nginx-kubernetes-gateway:edge

install-nkg:
	kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.6.2/standard-install.yaml
	sleep 30s # to allow webhook to get ready
	kubectl apply -f ../deploy/manifests/namespace.yaml
	kubectl create configmap njs-modules --from-file=../internal/nginx/modules/src/httpmatches.js -n nginx-gateway
	kubectl apply -f ../deploy/manifests/gatewayclass.yaml
	cat  ../deploy/manifests/nginx-gateway.yaml | sed "s|image: ghcr.io/nginxinc/nginx-kubernetes-gateway.*|image: nginx-kubernetes-gateway:edge|" | sed "s|imagePullPolicy: Always|imagePullPolicy: Never|" | kubectl apply -f -

run-tests-in-kind: update-test-kind-config ## Run tests in Kind
	docker run --network=kind --rm -v $(KIND_KUBE_CONFIG_FOLDER):/root/.kube $(PREFIX):$(TAG) \
		go test -v . -args --gateway-class=nginx

update-test-kind-config:
	sed -ir "s|server:.*|server: https://kind-control-plane:6443|" $(KIND_KUBE_CONFIG_FOLDER)/config