#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CORE_DIR="$ROOT_DIR/apps/core"

echo "== Validate apps/core tests =="
(cd "$CORE_DIR" && go test ./...)

echo
echo "== Smoke test apps/core healthz =="
chmod +x "$ROOT_DIR/scripts/verify-core-health.sh"
"$ROOT_DIR/scripts/verify-core-health.sh"

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
