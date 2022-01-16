BRANCH = `git rev-parse --abbrev-ref HEAD`
COMMIT = `git rev-parse --short HEAD`
# Image URL to use all building/pushing image targets
OPERATOR_IMG ?= rumia/rds-operator:$(BRANCH)
SIDECAR_IMG ?= rumia/rds-sidecar:$(BRANCH)
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true,preserveUnknownFields=false"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=rds-operator webhook paths="./..." output:crd:artifacts:config=assets/config/crd/bases output:dir=assets/config/rbac

generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...

##@ Deployment
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build assets/config/crd | kubectl apply -f -

uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build assets/config/crd | kubectl delete -f -

CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
controller-gen: ## Download controller-gen locally if necessary.
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.6.2)

KUSTOMIZE = $(shell pwd)/bin/kustomize
kustomize: ## Download kustomize locally if necessary.
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v4@v4.3.0)

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

#########################################################################################################################

sidecar-dev:
	go build -o sidecar cmd/sidecar/*.go

	docker build -t $(SIDECAR_IMG) \
	--build-arg Version=$(BRANCH) \
	--build-arg Commit=$(COMMIT) \
	-f assets/docker/sidecar/Dockerfile.dev .

sidecar:
	docker build -t $(SIDECAR_IMG) \
	--build-arg Version=$(BRANCH) \
	--build-arg Commit=$(COMMIT) \
	-f assets/docker/sidecar/Dockerfile .
operator:
	docker build -t $(OPERATOR_IMG) \
	--build-arg Version=$(BRANCH) \
	--build-arg Commit=$(COMMIT) \
	-f assets/docker/operator/Dockerfile .
yaml: manifests
	rm -rf release
	mkdir -p release/operator
	cd assets/config/operator && $(KUSTOMIZE) edit set image operator=rumia/rds-operator:${BRANCH}
	
	$(KUSTOMIZE) build assets/config/operator > release/operator/operator.yaml
	$(KUSTOMIZE) build assets/config/crd > release/operator/crd.yaml
	$(KUSTOMIZE) build assets/config/rbac > release/operator/rbac.yaml
	$(KUSTOMIZE) build assets/config/prometheus > release/operator/prometheus.yaml

	cp -r assets/examples release/examples

	sed -i 's/rumia\/rds-sidecar.*/rumia\/rds-sidecar:$(BRANCH)/g' release/examples/*.yaml

	rm -rf release/examples/kafka.yaml
	rm -rf release/operator/proxysql.yaml
	rm -rf release/operator/prometheus.yaml

	tar -czf release/yaml.tar.gz release/examples release/operator

gen: generate manifests install
	cd hack && ./update-codegen.sh
	
dev:
	go build -ldflags "-s -w -X github.com/hakur/rds-operator/pkg/types.Version=$(BRANCH) -X github.com/hakur/rds-operator/pkg/types.Commit=$(COMMIT)" -o rds-operator cmd/operator/main.go
	source hack/dev-env.sh && ./rds-operator

skaffold:
	go build -ldflags "-s -w -X github.com/hakur/rds-operator/pkg/types.Version=$(BRANCH) -X github.com/hakur/rds-operator/pkg/types.Commit=$(COMMIT)" -o rds-operator cmd/operator/main.go
	skaffold dev --port-forward=user,services --tail

# make relase BRANCH=${version tag} PUSH=true
release: yaml operator sidecar # make release pack
ifeq ($(PUSH),true)
	docker push $(SIDECAR_IMG)
	docker push $(OPERATOR_IMG)
endif
	