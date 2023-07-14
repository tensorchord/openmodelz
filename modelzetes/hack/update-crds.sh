#!/bin/bash

export controllergen="$GOPATH/bin/controller-gen"
export PKG=sigs.k8s.io/controller-tools/cmd/controller-gen@v0.7.0

if [ ! -e "$controllergen" ]; then
  echo "Getting $PKG"
  go install $PKG
fi

"$controllergen" \
  crd \
  schemapatch:manifests=./artifacts/crds \
  paths=./pkg/apis/modelzetes/v2alpha1 \
  output:dir=./artifacts/crds
