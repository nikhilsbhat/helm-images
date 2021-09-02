#! /bin/bash -e

function download_plugin() {
  osName=$(uname -s)
  DOWNLOAD_URL=$(curl --silent "https://api.github.com/repos/nikhilsbhat/helm-images/releases/latest" | grep -o "browser_download_url.*\_${osName}_x86_64.zip")

  DOWNLOAD_URL=${DOWNLOAD_URL//\"/}
  DOWNLOAD_URL=${DOWNLOAD_URL/browser_download_url: /}

  OUTPUT_BASENAME=helm-images
  OUTPUT_BASENAME_WITH_POSTFIX="$HELM_PLUGIN_DIR$OUTPUT_BASENAME.zip"

  if [ -z "$DOWNLOAD_URL" ]; then
    echo "Unsupported OS / architecture: ${osName}"
    exit 1
  fi

  #  echo "$DOWNLOAD_URL"
  if [[ -n $(command -v curl) ]]; then
    curl -L $DOWNLOAD_URL -o $OUTPUT_BASENAME_WITH_POSTFIX
  else
    echo "Need curl"
    exit -1
  fi

  echo $OUTPUT_BASENAME_WITH_POSTFIX
}

function install_plugin() {
  local HELM_PLUGIN_ARTIFACT_PATH=$1
  #  echo "HELM_PLUGIN_ARTIFACT_PATH: $HELM_PLUGIN_ARTIFACT_PATH"
  echo "Preparing to install into ${HELM_PLUGIN_DIR}"
  mkdir -p "$HELM_PLUGIN_DIR/helm-images/bin"
  tar -xvf "$HELM_PLUGIN_ARTIFACT_PATH" -C "$HELM_PLUGIN_DIR"helm-images
  mv "$HELM_PLUGIN_DIR"helm-images/helm-images "$HELM_PLUGIN_DIR"helm-images/bin
  rm -rf "$HELM_PLUGIN_ARTIFACT_PATH"
}

function install() {
  echo "Installing helm-images..."

  local ARTIFACT_PATH=$(download_plugin)
  set +e
  install_plugin "$ARTIFACT_PATH"
  local INSTALL_PLUGIN_STAT=$?
  set -e

  if [ ! $INSTALL_PLUGIN_STAT -eq 0 ]; then
    echo "installing helm plugin helm-images failed with error code: $INSTALL_PLUGIN_STAT"
    exit 1
  fi

  echo "helm-images is installed."
  echo
  echo "See https://github.com/nikhilsbhat/helm-images for help getting started."
}

install "$@"
