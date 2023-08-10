#!/usr/bin/env bash

set -x
set -o errexit
set -o nounset
set -o pipefail

kind=${1:-"build"}
deps=${2:-"remote"}

PROJECT_ROOT=$(realpath "$(dirname "${BASH_SOURCE[@]}")/..")
PROJECT=github.com/tensorchord/openmodelz/mdz

VERSION=${VERSION:-"v0.0.$(date +%Y%m%d)"}
BUILD_DATE=${BUILD_DATE:-""}
GIT_COMMIT=${GIT_COMMIT:-""}
GIT_TREE_STATE=${GIT_TREE_STATE:-""}
GIT_TAG=${GIT_TAG:-""}

CGO_ENABLED=${CGO_ENABLED:="0"}
GOOS=${GOOS:="$(go env GOOS)"}
GOARCH=${GOARCH:="$(go env GOARCH)"}

DASHBOARD_BUILD=${DASHBOARD_BUILD:-"debug"}

OUTPUT_DIR=${OUTPUT_DIR:-"${PROJECT_ROOT}"}
if [[ ${kind} == "debug" ]]; then
    OUTPUT_DIR="${OUTPUT_DIR}/debug-bin"
else
    OUTPUT_DIR="${OUTPUT_DIR}/bin"
fi

LDFLAGS="-s -w \
  -X ${PROJECT}/pkg/version.version=${VERSION} \
  -X ${PROJECT}/pkg/version.buildDate=${BUILD_DATE} \
  -X ${PROJECT}/pkg/version.gitCommit=${GIT_COMMIT} \
  -X ${PROJECT}/pkg/version.gitTreeState=${GIT_TREE_STATE}"
if [[ ${kind} == "release" ]]; then
    LDFLAGS="${LDFLAGS} -X ${PROJECT}/pkg/version.gitTag=${GIT_TAG}"
fi

DEBUG_ARGS=()
if [[ ${kind} == "debug" ]]; then
    DEBUG_ARGS=(-tags "${DASHBOARD_BUILD}" -gcflags="all=-N -l")
fi

MOD_FILE="${PROJECT_ROOT}/go.mod"
if [[ ${deps} == "local" ]]; then
    if ! grep "\.\./agent" "${MOD_FILE}"; then
        echo "replace github.com/tensorchord/openmodelz/agent => ../agent" >> "${MOD_FILE}"
        go mod tidy
    fi
else
    if grep "\.\./agent" "${MOD_FILE}"; then
        sed -i -e '/..agent/d' "${MOD_FILE}"
        go mod tidy
    fi
fi

CGO_ENABLED=${CGO_ENABLED} GOOS=${GOOS} GOARCH=${GOARCH} go build \
           "${DEBUG_ARGS[@]}" \
           -trimpath -v \
           -ldflags "${LDFLAGS}" \
           -o "${OUTPUT_DIR}/mdz" \
           "./cmd/mdz"
