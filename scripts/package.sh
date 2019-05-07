#!/usr/bin/env bash
set -eo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."
./scripts/install_tools.sh

PACKAGE_DIR=${PACKAGE_DIR:-"${PWD##*/}_$(openssl rand -hex 4)"}

full_path=$(realpath "$PACKAGE_DIR")
args=".bin/packager -uncached"

if [[ $1 == "-c" ]]; then #package as cached
    full_path="$full_path-cached"
    args=".bin/packager"
fi

eval "$args $full_path"
