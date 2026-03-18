#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
REPO_ROOT=$(CDPATH='' cd -- "${SCRIPT_DIR}/.." && pwd)
DIST_DIR="${REPO_ROOT}/dist"

if ! command -v helm >/dev/null 2>&1; then
  printf 'helm is required to package the plugin\n' >&2
  exit 1
fi

HELM_MAJOR=$(helm version --template '{{ .Version }}' | cut -d. -f1 | tr -d 'v')
if [ "${HELM_MAJOR}" -lt 4 ]; then
  printf 'helm 4 or newer is required to create a Helm 4 compliant plugin package\n' >&2
  exit 1
fi

if [ -z "${HELM_PLUGIN_KEY_NAME:-}" ]; then
  printf 'HELM_PLUGIN_KEY_NAME must be set\n' >&2
  exit 1
fi

if [ -z "${HELM_PLUGIN_KEYRING:-}" ]; then
  printf 'HELM_PLUGIN_KEYRING must be set\n' >&2
  exit 1
fi

mkdir -p "${DIST_DIR}"
rm -f "${DIST_DIR}"/*.tgz "${DIST_DIR}"/*.tgz.prov

helm plugin package "${REPO_ROOT}" \
  --destination "${DIST_DIR}" \
  --key "${HELM_PLUGIN_KEY_NAME}" \
  --keyring "${HELM_PLUGIN_KEYRING}"

printf 'Created Helm plugin package and provenance in %s\n' "${DIST_DIR}"
