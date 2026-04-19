#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CORE_DIR="$ROOT_DIR/apps/core"

echo "== Validate apps/core tests =="
(cd "$CORE_DIR" && go test ./...)

echo
echo "== Build apps/core container =="
docker build -f "$CORE_DIR/manifest/docker/Dockerfile" -t easymvp-core:verify "$CORE_DIR"

if command -v kubectl >/dev/null 2>&1; then
  echo
  echo "== Render kustomize overlays =="
  kubectl kustomize "$CORE_DIR/manifest/deploy/kustomize/overlays/develop" >/dev/null
  kubectl kustomize "$CORE_DIR/manifest/deploy/kustomize/overlays/production" >/dev/null
else
  echo
  echo "== Skip kustomize render =="
  echo "kubectl not found; skipping kubectl kustomize validation"
fi
