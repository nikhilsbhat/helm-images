#! /bin/bash -e

function download() {
  osName=$(uname -s)
  DOWNLOAD_URL=$(curl --silent "https://api.github.com/repos/nikhilsbhat/helm-images/releases/latest" | grep -o "browser_download_url.*\_${osName}_x86_64.tar.gz")

  DOWNLOAD_URL=${DOWNLOAD_URL//\"/}
  DOWNLOAD_URL=${DOWNLOAD_URL/browser_download_url: /}

  echo "download url: $DOWNLOAD_URL"
  OUTPUT_BASENAME=helm-images
  OUTPUT_BASENAME_WITH_POSTFIX=$HELM_PLUGIN_DIR/$OUTPUT_BASENAME.tar.gz

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
  HELM_TMP="/tmp/$PROJECT_NAME"
  mkdir -p "$HELM_TMP"
  tar xf "$PLUGIN_TMP_FILE" -C "$HELM_TMP"
  HELM_TMP_BIN="$HELM_TMP/diff/bin/diff"
  echo "Preparing to install into ${HELM_PLUGIN_DIR}"
  mkdir -p "$HELM_PLUGIN_DIR/bin"
  cp "$HELM_TMP_BIN" "$HELM_PLUGIN_DIR/bin"

}

function install() {
  echo "Installing helm-images..."

  local artifact_path=$(download)
  echo "$artifact_path"
  #  rm -rf bin && mkdir bin && tar -xvf $OUTPUT_BASENAME_WITH_POSTFIX -C bin >/dev/null && rm -f $OUTPUT_BASENAME_WITH_POSTFIX

  echo "helm-images is installed."
  echo
  echo "See https://github.com/nikhilsbhat/helm-images for help getting started."
}

install "$@"
