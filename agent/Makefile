# Copyright 2022 TensorChord Inc.
#
# The old school Makefile, following are required targets. The Makefile is written
# to allow building multiple binaries. You are free to add more targets or change
# existing implementations, as long as the semantics are preserved.
#
#   make              - default to 'build' target
#   make lint         - code analysis
#   make test         - run unit test (or plus integration test)
#   make build        - alias to build-local target
#   make build-local  - build local binary targets
#   make build-linux  - build linux binary targets
#   make container    - build containers
#   $ docker login registry -u username -p xxxxx
#   make push         - push containers
#   make clean        - clean up targets
#
# Not included but recommended targets:
#   make e2e-test
#
# The makefile is also responsible to populate project version information.
#

#
# Tweak the variables based on your project.
#

# This repo's root import path (under GOPATH).
ROOT := github.com/tensorchord/openmodelz/agent

# Target binaries. You can build multiple binaries for a single project.
TARGETS := agent

# Container image prefix and suffix added to targets.
# The final built images are:
#   $[REGISTRY]/$[IMAGE_PREFIX]$[TARGET]$[IMAGE_SUFFIX]:$[VERSION]
# $[REGISTRY] is an item from $[REGISTRIES], $[TARGET] is an item from $[TARGETS].
IMAGE_PREFIX ?= $(strip )
IMAGE_SUFFIX ?= $(strip )

# Container registries.
REGISTRY ?= ghcr.io/tensorchord

# Container registry for base images.
BASE_REGISTRY ?= docker.io
BASE_REGISTRY_USER ?= modelzai

# Disable CGO by default.
CGO_ENABLED ?= 0

#
# These variables should not need tweaking.
#

# It's necessary to set this because some environments don't link sh -> bash.
export SHELL := bash

# It's necessary to set the errexit flags for the bash shell.
export SHELLOPTS := errexit

PACKAGE_NAME := github.com/tensorchord/openmodelz/agent
GOLANG_CROSS_VERSION  ?= v1.17.6

# Project main package location (can be multiple ones).
CMD_DIR := ./cmd

# Project output directory.
OUTPUT_DIR := ./bin
DEBUG_DIR := ./debug-bin

# Build directory.
BUILD_DIR := ./build

# Current version of the project.
VERSION ?= $(shell git describe --match 'v[0-9]*' --always --tags --abbrev=0)
BUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_TAG=$(shell if [ -z "`git status --porcelain`" ]; then git describe --exact-match --tags HEAD 2>/dev/null; fi)
GIT_TREE_STATE=$(shell if [ -z "`git status --porcelain`" ]; then echo "clean" ; else echo "dirty"; fi)
GITSHA ?= $(shell git rev-parse --short HEAD)

# Track code version with Docker Label.
DOCKER_LABELS ?= git-describe="$(shell date -u +v%Y%m%d)-$(shell git describe --tags --always --dirty)"

# Golang standard bin directory.
GOPATH ?= $(shell go env GOPATH)
GOROOT ?= $(shell go env GOROOT)
BIN_DIR := $(GOPATH)/bin
GOLANGCI_LINT := $(BIN_DIR)/golangci-lint

# check if we need embed the dashboard
DASHBOARD_BUILD ?= debug

# Default golang flags used in build and test
# -mod=vendor: force go to use the vendor files instead of using the `$GOPATH/pkg/mod`
# -p: the number of programs that can be run in parallel
# -count: run each test and benchmark 1 times. Set this flag to disable test cache
export GOFLAGS ?= -count=1

#
# Define all targets. At least the following commands are required:
#

# All targets.
.PHONY: help lint test build container push addlicense debug debug-local build-local generate clean test-local addlicense-install release build-image

.DEFAULT_GOAL:=build

build: build-local  ## Build the release version

help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

debug: debug-local  ## Build the debug version

# more info about `GOGC` env: https://github.com/golangci/golangci-lint#memory-usage-of-golangci-lint
lint: $(GOLANGCI_LINT)  ## Lint GO code
	@$(GOLANGCI_LINT) run

$(GOLANGCI_LINT):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin

mockgen-install:
	go install github.com/golang/mock/mockgen@v1.6.0

addlicense-install:
	go install github.com/google/addlicense@latest

sqlc-install:
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# https://github.com/swaggo/swag/pull/1322, we should use master instead of latest for now.
swag-install:
	go install github.com/swaggo/swag/cmd/swag@v1.8.7

build-local:
	@for target in $(TARGETS); do                                                      \
	  CGO_ENABLED=$(CGO_ENABLED) go build -tags $(DASHBOARD_BUILD)  -trimpath -v -o $(OUTPUT_DIR)/$${target}     \
	    -ldflags "-s -w -X $(ROOT)/pkg/version.version=$(VERSION) -X $(ROOT)/pkg/version.buildDate=$(BUILD_DATE) -X $(ROOT)/pkg/version.gitCommit=$(GIT_COMMIT) -X $(ROOT)/pkg/version.gitTreeState=$(GIT_TREE_STATE)"                     \
	    $(CMD_DIR)/$${target};                                                         \
	done

# It is used by vscode to attach into the process.
debug-local:
	@for target in $(TARGETS); do                                                      \
	  CGO_ENABLED=$(CGO_ENABLED) go build -tags $(DASHBOARD_BUILD) -trimpath                                    \
	  	-v -o $(DEBUG_DIR)/$${target}                                                  \
	  	-gcflags='all=-N -l'                                                           \
	    $(CMD_DIR)/$${target};                                                         \
	done

addlicense: addlicense-install  ## Add license to GO code files
	addlicense -l mpl -c "TensorChord Inc." $$(find . -type f -name '*.go' | grep -v pkg/docs/docs.go)

test-local:
	@go test -tags=$(DASHBOARD_BUILD) -v -race -coverprofile=coverage.out ./...

test:  ## Run the tests
	@go test -tags=$(DASHBOARD_BUILD) -race -coverprofile=coverage.out ./...
	@go tool cover -func coverage.out | tail -n 1 | awk '{ print "Total coverage: " $$3 }'

clean:  ## Clean the outputs and artifacts
	@-rm -vrf ${OUTPUT_DIR}
	@-rm -vrf ${DEBUG_DIR}
	@-rm -vrf build dist .eggs *.egg-info

fmt: swag-install ## Run go fmt against code.
	go fmt ./...
	swag fmt

vet: ## Run go vet against code.
	go vet ./...

swag: swag-install
	swag init -g ./cmd/agent/main.go --parseDependency --output ./pkg/docs 

build-image: build-local
	docker build -t ${BASE_REGISTRY}/${BASE_REGISTRY_USER}/modelz-agent:dev -f Dockerfile ./bin
	docker push ${BASE_REGISTRY}/${BASE_REGISTRY_USER}/modelz-agent:dev

release:
	@if [ ! -f ".release-env" ]; then \
		echo "\033[91m.release-env is required for release\033[0m";\
		exit 1;\
	fi
	docker run \
		--rm \
		--privileged \
		-e CGO_ENABLED=1 \
		--env-file .release-env \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/$(PACKAGE_NAME) \
		-v `pwd`/sysroot:/sysroot \
		-w /go/src/$(PACKAGE_NAME) \
		goreleaser/goreleaser-cross:${GOLANG_CROSS_VERSION} \
		release --rm-dist

generate: mockgen-install sqlc-install swag
	@mockgen -source pkg/query/querier.go -destination pkg/query/mock/mock.go -package mock
	@mockgen -source pkg/runtime/runtime.go -destination pkg/runtime/mock/mock.go -package mock
	# @sqlc generate
