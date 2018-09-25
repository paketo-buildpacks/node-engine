#!/usr/bin/env bash
set -euo pipefail

cd "$( dirname "${BASH_SOURCE[0]}" )/.."

if [ ! -d .bin ]; then
  mkdir .bin
fi

export GOBIN=$PWD/.bin
export PATH=$GOBIN:$PATH

go install github.com/onsi/ginkgo/ginkgo

host=$([ $(uname -s) == 'Darwin' ] &&  printf "macos" || printf "linux")
version=$(curl --silent "https://api.github.com/repos/buildpack/pack/releases/latest" | jq -r .tag_name)
wget --quiet "https://github.com/buildpack/pack/releases/download/$version/pack-$host" -O $GOBIN/pack && chmod +x $GOBIN/pack
