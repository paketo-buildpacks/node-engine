#!/usr/bin/env bash
set -euo pipefail

cd "$( dirname "${BASH_SOURCE[0]}" )/.."

if [ ! -d .bin ]; then
  mkdir .bin
fi

export GOBIN=$PWD/.bin
export PATH=$GOBIN:$PATH

go install github.com/onsi/ginkgo/ginkgo
