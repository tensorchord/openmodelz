#!/usr/bin/env bash

# copied from: https://github.com/weaveworks/flagger/tree/master/hack

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(git rev-parse --show-toplevel)/modelzetes

echo ">> SCRIPT_ROOT ${SCRIPT_ROOT}"

GET_PKG_LOCATION() {
  pkg_name="${1:-}"

  pkg_location="$(go list -m -f '{{.Dir}}' "${pkg_name}" 2>/dev/null)"
  if [ "${pkg_location}" = "" ]; then
    echo "${pkg_name} is missing. Running 'go mod download'."

    go mod download
    pkg_location=$(go list -m -f '{{.Dir}}' "${pkg_name}")
  fi
  echo "${pkg_location}"
}

# Grab code-generator version from go.sum
CODEGEN_PKG="$(GET_PKG_LOCATION "k8s.io/code-generator")"
echo ">> Using ${CODEGEN_PKG}"

# Grab openapi-gen version from go.mod
OPENAPI_PKG="$(GET_PKG_LOCATION 'k8s.io/kube-openapi')"
echo ">> Using ${OPENAPI_PKG}"

echo ">> Using ${CODEGEN_PKG}"

# code-generator does work with go.mod but makes assumptions about
# the project living in `$GOPATH/src`. To work around this and support
# any location; create a temporary directory, use this as an output
# base, and copy everything back once generated.
TEMP_DIR=$(mktemp -d)
cleanup() {
    echo ">> Removing ${TEMP_DIR}"
    rm -rf ${TEMP_DIR}
}
trap "cleanup" EXIT SIGINT

echo ">> Temporary output directory ${TEMP_DIR}"

# Ensure we can execute.
chmod +x ${CODEGEN_PKG}/generate-groups.sh

${CODEGEN_PKG}/generate-groups.sh all \
    github.com/tensorchord/openmodelz/modelzetes/pkg/client github.com/tensorchord/openmodelz/modelzetes/pkg/apis \
    modelzetes:v2alpha1 \
    --output-base "${TEMP_DIR}" \
    --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.go.txt

# Copy everything back.
cp -r "${TEMP_DIR}/github.com/tensorchord/openmodelz/modelzetes/." "${SCRIPT_ROOT}/"
