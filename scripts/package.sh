#!/usr/bin/env bash
set -euo pipefail

TARGET_OS=${1:-linux}

cd "$( dirname "${BASH_SOURCE[0]}" )/.."

echo "Target OS is $TARGET_OS"
echo -n "Creating buildpack directory..."
bp_dir=/tmp/"${PWD##*/}"_$(openssl rand -hex 12)
mkdir $bp_dir
echo "done"

echo -n "Copying buildpack.toml..."
cp buildpack.toml $bp_dir/buildpack.toml
echo "done"

# TODO: Update list of built binaries as they are written
for b in detect build; do
    echo -n "Building $b..."
    GOOS=$TARGET_OS go build -o $bp_dir/bin/$b ./cmd/$b
    echo "done"
done
echo "Buildpack packaged into: $bp_dir"
