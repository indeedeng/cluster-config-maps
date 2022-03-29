#!/bin/bash

set -euo pipefail

rm -rf deploy/
git checkout main -- deploy/
helm package deploy/charts/cluster-config-maps
helm repo index . --url https://indeedeng.github.io/cluster-config-maps --merge index.yaml
