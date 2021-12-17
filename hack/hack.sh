#!/bin/bash
set -euo pipefail

cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")"

kind delete cluster --name ccm-cluster
kind create cluster --name ccm-cluster --config=cluster.yaml
(
  cd ..
  rm -rf bin/
  VERSION=dev IMAGE_REGISTRY=csi-ccm-plugin make docker.build
  make manifests
)
kind load docker-image csi-ccm-plugin:dev --name ccm-cluster
kubectl cluster-info --context kind-ccm-cluster
kubectl apply -f ../bin/deploy/manifests/*.yaml
kubectl apply -f deploy/demo.yaml
