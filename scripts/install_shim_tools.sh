#!/bin/bash
set -euo pipefail

cd "$( dirname "${BASH_SOURCE[0]}" )/.."
source .envrc
go mod download

if [ ! -f .bin/ginkgo ]; then
  go get -u github.com/onsi/ginkgo/ginkgo
fi
if [ ! -f .bin/buildpack-packager ]; then
  go install github.com/cloudfoundry/libbuildpack/packager/buildpack-packager
fi

if [ ! -f .bin/cnb2cf ]; then
    go install github.com/cloudfoundry/cnb2cf/cmd/cnb2cf
    ./scripts/build.sh
    if [ -f .bin/statik ]; then
        rm .bin/statik
    fi
fi

go mod tidy
