VERSION ?= 0.0.4

# check if we are using MacOS or LINUX and use that to determine the sed command
UNAME_S := $(shell uname -s)
SED=$(shell which gsed || which sed)

# CHANNELS define the bundle channels used in the bundle.
# Add a new line here if you would like to change its default config. (E.g CHANNELS = "candidate,fast,stable")
# To re-generate a bundle for other specific channels without changing the standard setup, you can:
# - use the CHANNELS as arg of the bundle target (e.g make bundle CHANNELS=candidate,fast,stable)
# - use environment variables to overwrite this value (e.g export CHANNELS="candidate,fast,stable")
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif

# DEFAULT_CHANNEL defines the default channel used in the bundle.
# Add a new line here if you would like to change its default config. (E.g DEFAULT_CHANNEL = "stable")
# To re-generate a bundle for any other default channel without changing the default setup, you can:
# - use the DEFAULT_CHANNEL as arg of the bundle target (e.g make bundle DEFAULT_CHANNEL=stable)
# - use environment variables to overwrite this value (e.g export DEFAULT_CHANNEL="stable")
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# IMAGE_TAG_BASE defines the docker.io namespace and part of the image name for remote images.
# This variable is used to construct full image tags for bundle and catalog images.
#
# For example, running 'make bundle-build bundle-push catalog-build catalog-push' will build and push both
IMAGE_TAG_BASE ?= docker.io/openvirtualcluster/operator

# BUNDLE_IMG defines the image:tag used for the bundle.
# You can use it as an arg. (E.g make bundle-build BUNDLE_IMG=<some-registry>/<project-name-bundle>:<tag>)
BUNDLE_IMG ?= $(IMAGE_TAG_BASE)-bundle:v$(VERSION)

# BUNDLE_GEN_FLAGS are the flags passed to the operator-sdk generate bundle command
BUNDLE_GEN_FLAGS ?= -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS)

# USE_IMAGE_DIGESTS defines if images are resolved via tags or digests
# You can enable this value if you would like to use SHA Based Digests
# To enable set flag to true
USE_IMAGE_DIGESTS ?= false
ifeq ($(USE_IMAGE_DIGESTS), true)
	BUNDLE_GEN_FLAGS += --use-image-digests
endif

# Image URL to use all building/pushing image targets
IMG ?= $(IMAGE_TAG_BASE):v$(VERSION)
LATEST_IMG ?= $(IMAGE_TAG_BASE):latest
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.26.0

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build build-helm-chart

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: install-flux-prereq
install-flux-prereq: ## Install the fluxcd if not preset
    which flux || curl -s https://fluxcd.io/install.sh | sudo bash

.PHONY: install-fluxcd-controllers
install-fluxcd-controllers: install-flux-prereq ## Install the fluxcd controllers.
	flux install --namespace=flux-system --components="source-controller,helm-controller" --network-policy=false --insecure-skip-tls-verify

.PHONE: install-fluxcd-controllers-with-toleration
install-fluxcd-controllers-with-toleration: install-flux-prereq ## Install the fluxcd controllers with toleration.
	flux install --namespace=flux-system --components="source-controller,helm-controller" --toleration-keys="sandbox.gke.io/runtime" --network-policy=false --insecure-skip-tls-verify

.PHONY: start-test-k3d
start-test-k3d: ## Start a k3d cluster for testing.
	k3d cluster create basic agents=1
	$(MAKE) install-fluxcd-controllers

.PHONY: start-test-minikube
start-test-minikube: ## Start a minikube cluster for testing.
	minikube start --addons default-storageclass,storage-provisioner --driver=docker
	kubectl taint nodes minikube testkey- || true
	kubectl label nodes minikube testkey- || true
	$(MAKE) install-fluxcd-controllers

.PHONY: stop-test-minikube
stop-test-minikube: ## Stop the minikube cluster for testing.
	minikube stop

.PHONY: start-test-minikube-tainted
start-test-minikube-tainted: ## Start a minikube cluster with a tainted node for testing.
	minikube start --addons default-storageclass,storage-provisioner --driver=docker
	sh ./hack/minikube/patch-pod-tolerations.sh
	kubectl taint nodes minikube sandbox.gke.io/runtime=gvisor:NoSchedule || true
	kubectl label nodes minikube sandbox.gke.io/runtime=gvisor || true
	$(MAKE) install-fluxcd-controllers-with-toleration
	sh ./hack/minikube/patch-workload-tolerations.sh

.PHONY : stop-test-k3d
stop-test-k3d: ## Stop the k3d cluster for testing.
	k3d cluster delete basic

.PHONY: start-test-k3d-tainted
start-test-k3d-tainted: ## Start a k3d cluster with a tainted node for testing.
	k3d cluster create tainted --agents=1 --k3s-arg="--kubelet-arg=node-labels=testkey=testvalue@agent:0" --k3s-arg="--kubelet-arg=taints=testkey=testvalue:NoSchedule@agent:0"
	$(MAKE) install-fluxcd-controllers

.PHONY : stop-test-k3d-tainted
stop-test-k3d-tainted: ## Stop the k3d cluster with a tainted node for testing.
	k3d cluster delete tainted

##@ Test

.PHONY: test
test: test-e2e-without-cluster

.PHONY: test-e2e-without-cluster
test-e2e-without-cluster: manifests generate fmt vet envtest ## Run test.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test ./... -coverprofile=coverage.txt

.PHONY: test-e2e-with-cluster
test-e2e-with-cluster: manifests generate fmt vet envtest ## Run test.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" ENVTEST_REMOTE=true go test ./... -coverprofile=coverage.txt -v

.PHONY: test-e2e-perf-with-cluster
test-e2e-perf-with-cluster:
	./hack/e2e/perf/main.sh

.PHONY: test-e2e-with-cluster-local
test-e2e-with-cluster-local: start-test-minikube test-e2e-with-cluster ## Run test.

.PHONY: test-e2e-with-tainted-cluster
test-e2e-with-tainted-cluster: manifests generate fmt vet envtest ## Run test.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" ENVTEST_REMOTE=true E2E_ARG_IS_TAINTED=true go test ./... -coverprofile=coverage.txt -v

.PHONY: test-e2e-with-tainted-cluster-local
test-e2e-with-tainted-cluster-local: start-test-minikube-tainted test-e2e-with-tainted-cluster ## Run test.

##@ Build

.PHONY: build
build: manifests generate fmt vet ## Build manager binary.
	go build -o bin/manager cmd/main.go

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run ./cmd/main.go

# If you wish built the manager image targeting other platforms you can use the --platform flag.
# (i.e. docker build --platform linux/arm64 ). However, you must enable docker buildKit for it.
# More info: https://docs.docker.com/develop/develop-images/build_enhancements/
.PHONY: docker-build
docker-build: ## Build docker image with the manager.
	docker build -t ${IMG} -t ${LATEST_IMG} .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	docker push ${IMG}
	docker push ${LATEST_IMG}

# PLATFORMS defines the target platforms for  the manager image be build to provide support to multiple
# architectures. (i.e. make docker-buildx IMG=myregistry/mypoperator:0.0.1). To use this option you need to:
# - able to use docker buildx . More info: https://docs.docker.com/build/buildx/
# - have enable BuildKit, More info: https://docs.docker.com/develop/develop-images/build_enhancements/
# - be able to push the image for your registry (i.e. if you do not inform a valid value via IMG=<myregistry/image:<tag>> then the export will fail)
# To properly provided solutions that supports more than one platform you should use this option.
PLATFORMS ?= linux/arm64,linux/amd64,linux/s390x,linux/ppc64le
.PHONY: docker-buildx
docker-buildx: ## Build and push docker image for the manager for cross-platform support
	# copy existing Dockerfile and insert --platform=${BUILDPLATFORM} into Dockerfile.cross, and preserve the original Dockerfile
	$(SED) -e '1 s/\(^FROM\)/FROM --platform=\$$\{BUILDPLATFORM\}/; t' -e ' 1,// s//FROM --platform=\$$\{BUILDPLATFORM\}/' Dockerfile > Dockerfile.cross
	- docker buildx create --name project-v3-builder
	docker buildx use project-v3-builder
	- docker buildx build --push --platform=$(PLATFORMS) --tag ${IMG} -f Dockerfile.cross .
	- docker buildx rm project-v3-builder
	rm Dockerfile.cross

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd src/config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: build-helm-chart
build-helm-chart: manifests generate fmt vet kustomize
	# Generate the combined manager.yaml
	$(KUSTOMIZE) build config/default > chart/templates/manager.yaml
	$(SPLIT_YAML)
	rm -f chart/templates/manager.yaml
	rm -f chart/templates/manager_namespace*

	# update the crd
	$(KUSTOMIZE) build config/crd > chart/templates/vclusters.openvirtualcluster.dev_customresourcedefinition.yaml
	# Remove existing controller-gen annotation
	$(SED) -i'' -e '/annotations:/,/^[[:space:]]*[^[:space:]]/d' chart/templates/vclusters.openvirtualcluster.dev_customresourcedefinition.yaml
	$(SED) -i'' -e 's/labels:/labels: {{ include "common.labels.standard" . | nindent 4 }}/' chart/templates/vclusters.openvirtualcluster.dev_customresourcedefinition.yaml
	$(SED) -i'' -e '/metadata:/a\
	\  annotations:\
	\n    meta.helm.sh/release-name: {{ .Release.Name }}\
	\n    meta.helm.sh/release-namespace: {{ .Release.Namespace }}\
	\n    controller-gen.kubebuilder.io/version: v0.13.0\
	\n  labels:\
	\n    app.kubernetes.io/managed-by: Helm' chart/templates/vclusters.openvirtualcluster.dev_customresourcedefinition.yaml

	# update chart versions
	yq e -i '.version = "${VERSION}"' chart/Chart.yaml
	yq e -i '.appVersion = "v${VERSION}"' chart/Chart.yaml
	yq e -i '.image.tag = "v${VERSION}"' chart/values.yaml

	# set metadata.namespace and update metadata.name in all templates
	for file in chart/templates/manager_*.yaml; do \
		if yq e 'has("metadata")' $$file; then \
			yq e -i '.metadata.namespace = "{{ .Release.Namespace }}"' $$file; \
			if yq e '.metadata.name | test("^openvirtualcluster")' $$file; then \
				yq e -i '.metadata.name |= sub("^openvirtualcluster", "{{ include \"common.names.fullname\" . }}")' $$file; \
			fi; \
		fi; \
	done

	# Extract the service account name from manager_serviceaccount
	SERVICE_ACCOUNT_NAME="$(shell yq e '.metadata.name' chart/templates/manager_serviceaccount_openvirtualcluster-controller-manager.yaml)" \
	export SERVICE_ACCOUNT_NAME; \
	# Update the service account name in the deployment and bindings
	yq e -i '(.spec.template.spec.serviceAccountName) = strenv(SERVICE_ACCOUNT_NAME)' chart/templates/manager_deployment_kube-rbac-proxy.yaml
	for file in chart/templates/manager_rolebinding*.yaml chart/templates/manager_clusterrolebinding*.yaml; do \
		echo $$file > /dev/null; \
        yq e -i '.subjects[] |= (select(.kind == "ServiceAccount") | .name = strenv(SERVICE_ACCOUNT_NAME) | .namespace = "{{ .Release.Namespace }}")' $$file; \
    done
	# Update the image tag in the manager_deployment_kube-rbac-proxy.yaml
	sed -i '' -E "s|image: controller:latest|image: 'docker.io/openvirtualcluster/operator:{{ .Values.image.tag }}'|" "chart/templates/manager_deployment_kube-rbac-proxy.yaml"


.PHONY: helm-lint
helm-lint: ## Lint the helm chart.
	(cd ./chart && helm dep update .)
	helm lint ./chart --with-subcharts

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
SPLIT_YAML ?= hack/split_yaml.sh

## Tool Versions
KUSTOMIZE_VERSION ?= v5.2.1
CONTROLLER_TOOLS_VERSION ?= v0.13.0

KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary. If wrong version is installed, it will be removed before downloading.
$(KUSTOMIZE): $(LOCALBIN)
	@if test -x $(LOCALBIN)/kustomize && ! $(LOCALBIN)/kustomize version | grep -q $(KUSTOMIZE_VERSION); then \
		echo "$(LOCALBIN)/kustomize version is not expected $(KUSTOMIZE_VERSION). Removing it before installing."; \
		rm -rf $(LOCALBIN)/kustomize; \
	fi
	test -s $(LOCALBIN)/kustomize || { curl -Ss $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN); }

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary. If wrong version is installed, it will be overwritten.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen && $(LOCALBIN)/controller-gen --version | grep -q $(CONTROLLER_TOOLS_VERSION) || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	@test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

.PHONY: bundle
bundle: manifests kustomize ## Generate bundle manifests and metadata, then validate generated files.
	operator-sdk generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	$(KUSTOMIZE) build config/manifests | operator-sdk generate bundle $(BUNDLE_GEN_FLAGS)
	operator-sdk bundle validate ./bundle

.PHONY: bundle-build
bundle-build: ## Build the bundle image.
	docker build -f bundle.Dockerfile -t $(BUNDLE_IMG) .

.PHONY: bundle-push
bundle-push: ## Push the bundle image.
	$(MAKE) docker-push IMG=$(BUNDLE_IMG)

.PHONY: opm
OPM = ./bin/opm
opm: ## Download opm locally if necessary.
ifeq (,$(wildcard $(OPM)))
ifeq (,$(shell which opm 2>/dev/null))
	@{ \
	set -e ;\
	mkdir -p $(dir $(OPM)) ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(OPM) https://github.com/operator-framework/operator-registry/releases/download/v1.23.0/$${OS}-$${ARCH}-opm ;\
	chmod +x $(OPM) ;\
	}
else
OPM = $(shell which opm)
endif
endif

# A comma-separated list of bundle images (e.g. make catalog-build BUNDLE_IMGS=example.com/operator-bundle:v0.1.0,example.com/operator-bundle:v0.2.0).
# These images MUST exist in a registry and be pull-able.
BUNDLE_IMGS ?= $(BUNDLE_IMG)

# The image tag given to the resulting catalog image (e.g. make catalog-build CATALOG_IMG=example.com/operator-catalog:v0.2.0).
CATALOG_IMG ?= $(IMAGE_TAG_BASE)-catalog:v$(VERSION)

# Set CATALOG_BASE_IMG to an existing catalog image tag to add $BUNDLE_IMGS to that image.
ifneq ($(origin CATALOG_BASE_IMG), undefined)
FROM_INDEX_OPT := --from-index $(CATALOG_BASE_IMG)
endif

# Build a catalog image by adding bundle images to an empty catalog using the operator package manager tool, 'opm'.
# This recipe invokes 'opm' in 'semver' bundle add mode. For more information on add modes, see:
# https://github.com/operator-framework/community-operators/blob/7f1438c/docs/packaging-operator.md#updating-your-existing-operator
.PHONY: catalog-build
catalog-build: opm ## Build a catalog image.
	$(OPM) index add --container-tool docker --mode semver --tag $(CATALOG_IMG) --bundles $(BUNDLE_IMGS) $(FROM_INDEX_OPT)

# Push the catalog image.
.PHONY: catalog-push
catalog-push: ## Push a catalog image.
	$(MAKE) docker-push IMG=$(CATALOG_IMG)
