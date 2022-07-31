#! /bin/bash -e

function verlte() {
    [  "$1" = "`echo -e "$1\n$2" | sort -V | head -n1`" ]
}

function verlt() {
    [ "$1" = "$2" ] && return 1 || verlte $1 $2
}

function isOld() {
    verlte $1 $2 && echo "yes" || echo "no"
}

function download_plugin() {
  osName=$(uname -s)
  osArch=$(uname -m)

  OUTPUT_BASENAME=helm-images
  version=$(grep version "$HELM_PLUGIN_DIR/plugin.yaml" | cut -d'"' -f2)
  old=$(isOld "$version" "0.0.5")
  if [ "$old" == "yes" ]; then
    DOWNLOAD_URL="https://github.com/nikhilsbhat/helm-images/releases/download/v$version/helm-images_${version}_${osName}_${osArch}.zip"
    OUTPUT_BASENAME_WITH_POSTFIX="$HELM_PLUGIN_DIR$OUTPUT_BASENAME.zip"
  else
    DOWNLOAD_URL="https://github.com/nikhilsbhat/helm-images/releases/download/v$version/helm-images_${version}_${osName}_${osArch}.tar.gz"
    OUTPUT_BASENAME_WITH_POSTFIX="$HELM_PLUGIN_DIR$OUTPUT_BASENAME.tar.gz"
  fi


  if [ -z "$DOWNLOAD_URL" ]; then
    echo "Unsupported OS / architecture: ${osName}/${osArch}"
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
  local PROJECT_NAME="helm-images"
  local HELM_PLUGIN_TEMP_PATH="/tmp/$PROJECT_NAME"

  rm -rf "$HELM_PLUGIN_TEMP_PATH"

  echo "Preparing to install into ${HELM_PLUGIN_DIR}"
  mkdir -p "$HELM_PLUGIN_TEMP_PATH" && tar -xvf "$HELM_PLUGIN_ARTIFACT_PATH" -C "$HELM_PLUGIN_TEMP_PATH"
  mkdir -p "$HELM_PLUGIN_DIR/bin"
  mv "$HELM_PLUGIN_TEMP_PATH"/helm-images "$HELM_PLUGIN_DIR/bin/helm-images"
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

  echo
  echo "helm-images is installed."
  echo
  "${HELM_PLUGIN_DIR}"/bin/helm-images -h
  echo
  echo "See https://github.com/nikhilsbhat/helm-images#readme for more information on getting started."
}

install "$@"
