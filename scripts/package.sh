#!/usr/bin/env bash
set -euo pipefail

TARGET_OS=${1:-linux}

cd "$( dirname "${BASH_SOURCE[0]}" )/.."

echo "Target OS is $TARGET_OS"
echo -n "Creating buildpack directory..."
bp_dir="${PWD##*/}"_$(openssl rand -hex 12)
mkdir $bp_dir
echo "done"

echo -n "Copying buildpack.toml..."
cp buildpack.toml $bp_dir/buildpack.toml
echo "done"

if [ "${BP_REWRITE_HOST:-}" != "" ]; then
    sed -i -e "s|^uri = \"https:\/\/buildpacks\.cloudfoundry\.org\(.*\)\"$|uri = \"http://$BP_REWRITE_HOST\1\"|g" "$bp_dir/buildpack.toml"
fi

for b in $(ls cmd); do
    echo -n "Building $b..."
    GOOS=$TARGET_OS go build -o $bp_dir/bin/$b ./cmd/$b
    echo "done"
done

fullPath=$(realpath "$bp_dir")
echo "Buildpack packaged into: $fullPath"

buildpack_name="$(basename `pwd`)"

pushd $bp_dir
    tar czvf "../$buildpack_name.tgz" *
popd
