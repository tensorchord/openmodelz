TAG?=dev
SERVER?=docker.io
OWNER?=tensorchord
IMAGE := ingress-operator
# This repo's root import path (under GOPATH).
ROOT := github.com/tensorchord/openmodelz/ingress-operator

# Target binaries. You can build multiple binaries for a single project.
TARGETS := ingress-operator
# Disable CGO by default.
CGO_ENABLED ?= 0
# Project main package location (can be multiple ones).
CMD_DIR := ./
OUTPUT_DIR := ./

.GIT_COMMIT=$(shell git rev-parse HEAD)
.GIT_VERSION=$(shell git describe --tags 2>/dev/null || echo "$(.GIT_COMMIT)")
.GIT_UNTRACKEDCHANGES := $(shell git status --porcelain --untracked-files=no)
ifneq ($(.GIT_UNTRACKEDCHANGES),)
	.GIT_COMMIT := $(.GIT_COMMIT)-dirty
endif

# docker manifest command will work with Docker CLI 18.03 or newer
# but for now it's still experimental feature so we need to enable that
export DOCKER_CLI_EXPERIMENTAL=enabled

TOOLS_DIR := .tools

GOPATH := $(shell go env GOPATH)
CODEGEN_VERSION := $(shell hack/print-codegen-version.sh)
CODEGEN_PKG := $(GOPATH)/pkg/mod/k8s.io/code-generator@${CODEGEN_VERSION}



ARCH?=linux/amd64
MULTIARCH?=linux/amd64,linux/arm/v7,linux/arm64

$(TOOLS_DIR)/code-generator.mod: go.mod
	@echo "syncing code-generator tooling version"
	@cd $(TOOLS_DIR) && go mod edit -require "k8s.io/code-generator@${CODEGEN_VERSION}"

${CODEGEN_PKG}: $(TOOLS_DIR)/code-generator.mod
	@echo "(re)installing k8s.io/code-generator-${CODEGEN_VERSION}"
	@cd $(TOOLS_DIR) && go mod download -modfile=code-generator.mod

.PHONY: build
build: build-local
	@echo "building  $(SERVER)/$(OWNER)/$(IMAGE):$(TAG)-$(.GIT_COMMIT)"
	@docker build \
	--build-arg VERSION=$(.GIT_VERSION) \
	--build-arg GIT_COMMIT=$(.GIT_COMMIT) \
	-t  $(SERVER)/$(OWNER)/$(IMAGE):$(TAG)-$(.GIT_COMMIT) .
	@rm -rf ingress-operator
	@docker push $(SERVER)/$(OWNER)/$(IMAGE):$(TAG)-$(.GIT_COMMIT) 

.PHONY: build-local
build-local:
	@for target in $(TARGETS); do                                                      \
		CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) go build -tags $(DASHBOARD_BUILD)  -trimpath -v -o $(OUTPUT_DIR)/$${target}     \
		-ldflags "-s -w -X $(ROOT)/pkg/version.version=$(VERSION) -X $(ROOT)/pkg/version.buildDate=$(BUILD_DATE) -X $(ROOT)/pkg/version.gitCommit=$(GIT_COMMIT) -X $(ROOT)/pkg/version.gitTreeState=$(GIT_TREE_STATE)"                     \
		$(CMD_DIR)/;                                                         \
	done

.PHONY: build-buildx
build-buildx:
	@echo  $(SERVER)/$(OWNER)/$(IMAGE):$(TAG) && \
	docker buildx create --use --name=multiarch --node=multiarch && \
	docker buildx build \
		--push \
		--platform $(ARCH) \
		--build-arg VERSION=$(.GIT_VERSION) \
		--build-arg GIT_COMMIT=$(.GIT_COMMIT) \
		--tag  $(SERVER)/$(OWNER)/$(IMAGE):$(TAG) \
		.

.PHONY: build-buildx-all
build-buildx-all:
	@echo  "build $(SERVER)/$(OWNER)/$(IMAGE):$(TAG) for $(MULTIARCH)"
	@docker buildx create --use --name=multiarch --node=multiarch && \
	docker buildx build \
		--platform $(MULTIARCH) \
		--output "type=image,push=false" \
		--build-arg VERSION=$(.GIT_VERSION) \
		--build-arg GIT_COMMIT=$(.GIT_COMMIT) \
		--tag $(SERVER)/$(OWNER)/$(IMAGE):$(TAG) \
		.

.PHONY: publish-buildx-all
publish-buildx-all:
	@echo  "build and publish $(SERVER)/$(OWNER)/$(IMAGE):$(TAG) for $(MULTIARCH)"
	@docker buildx create --use --name=multiarch --node=multiarch && \
	docker buildx build \
		--platform $(MULTIARCH) \
		--push=true \
		--build-arg VERSION=$(.GIT_VERSION) \
		--build-arg GIT_COMMIT=$(.GIT_COMMIT) \
		--tag $(SERVER)/$(OWNER)/$(IMAGE):$(TAG) \
		.

.PHONY: test
test:
	go test -v ./...

.PHONY: verify-codegen
verify-codegen: ${CODEGEN_PKG}
	./hack/verify-codegen.sh

.PHONY: update-codegen
update-codegen: ${CODEGEN_PKG}
	./hack/update-codegen.sh
