#!/usr/bin/env bash
set -eo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."
./scripts/install_tools.sh

PACKAGE_DIR=${PACKAGE_DIR:-"${PWD##*/}_$(openssl rand -hex 4)"}

full_path=$(realpath "$PACKAGE_DIR")
args=".bin/packager -uncached"

if [ $1 == "-c" ] || [ $2 == "-c" ]; then #package as cached
    full_path="$full_path-cached"
    args=".bin/packager"
fi

if [ $1 == "-a" ] || [ $2 == "-a" ]; then #package as archive
    args="${args} -archive"
fi
eval "$args $full_path"

if [[ -n "$BP_REWRITE_HOST" ]]; then
    sed -i '' -e "s|^uri = \"https:\/\/buildpacks\.cloudfoundry\.org\(.*\)\"$|uri = \"http://$BP_REWRITE_HOST\1\"|g" "$full_path/buildpack.toml"
fi
