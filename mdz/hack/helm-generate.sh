#/usr/bin/env bash

set -e

cd "$(dirname "$0")/.."

# Create the single YAML file from helm chart.
helm template ../charts/openmodelz --namespace openmodelz > pkg/server/openmodelz.yaml
cd - > /dev/null
