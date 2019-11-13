#!/usr/bin/env bash
set -eo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."
./scripts/install_tools.sh

PACKAGE_DIR=${PACKAGE_DIR:-"${PWD##*/}_$(openssl rand -hex 4)"}

while getopts "cv:" arg
do
    case $arg in
    c) offline=true;;
    v) version="${OPTARG}";;
    esac
done

if [[ -z "$version" ]]; then #version not provided, use latest git tag
    git_tag=$(git describe --abbrev=0 --tags)
    version=${git_tag:1}
fi

extra_args=""

if [[ -n "${offline}" ]]; then
    PACKAGE_DIR="${PACKAGE_DIR}-cached"
    extra_args+="--offline"
fi

.bin/jam pack \
    --buildpack "$(pwd)/buildpack.toml" \
    --version "${version}" \
    --output "${PACKAGE_DIR}" \
    ${extra_args}
