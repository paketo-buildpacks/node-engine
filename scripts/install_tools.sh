#!/usr/bin/env bash
set -euo pipefail

PACK_VERSION=latest

usage() {
    echo "Usage:   install_tools.sh <version: optional>"
    echo "Example: install_tools.sh 0.0.9"
    exit 0
}

if [ "$#" -gt 1  ]; then
    usage
fi

if [[ "$#" -eq 1 && $1 == "-h"  ]]; then
    usage
fi

if [ "$#" -eq 1  ]; then
    PACK_VERSION="$1"
fi

install_pack() {
    OS=$(uname -s)

    if [[ $OS == "Darwin" ]]; then
        OS="macos"
    elif [[ $OS == "Linux" ]]; then
        OS="linux"
    else
        echo "Unsupported operating system"
        exit 1
    fi

    if [ "$PACK_VERSION" != "latest" ]; then 
        echo "Installing pack $PACK_VERSION" 

        PACK_ARTIFACT=pack-$PACK_VERSION-$OS.tar.gz
        ARTIFACT_URL="https://github.com/buildpack/pack/releases/download/v$PACK_VERSION/$PACK_ARTIFACT"
        expand $ARTIFACT_URL
        return 0
    fi

    if [[ $OS == "macos" ]]; then
        ARTIFACT_URL=$(curl -s https://api.github.com/repos/buildpack/pack/releases/latest |   jq --raw-output '.assets[1] | .browser_download_url')
    else
        ARTIFACT_URL=$(curl -s https://api.github.com/repos/buildpack/pack/releases/latest |   jq --raw-output '.assets[0] | .browser_download_url')
    fi

    expand $ARTIFACT_URL
}

expand() {
    PACK_ARTIFACT=$(echo $1 | sed "s/.*\///")
    PACK_VERSION=v$(echo $PACK_ARTIFACT | sed 's/pack-//' | sed 's/-.*//')

    if [[ ! -f .bin/pack ]]; then
        echo "Installing Pack"
    elif [[ "$(.bin/pack version | sed 's/VERSION: //' | cut -d ' ' -f 1)" != *$PACK_VERSION* ]]; then
        rm .bin/pack
        echo "Updating Pack"
    else
        echo "Version $PACK_VERSION of pack is already installed"
        return 0
    fi
    wget $ARTIFACT_URL
    tar xzvf $PACK_ARTIFACT -C .bin
    rm $PACK_ARTIFACT
}

configure_pack() {
    pack add-stack org.cloudfoundry.stacks.cflinuxfs3 \
        --build-image cfbuildpacks/cflinuxfs3-cnb-experimental:build \
        --run-image cfbuildpacks/cflinuxfs3-cnb-experimental:run || echo "Ignoring add stack error"
}

cd "$( dirname "${BASH_SOURCE[0]}" )/.."

mkdir -p .bin
export PATH=$(pwd)/.bin:$PATH

install_pack
configure_pack
