#!/usr/bin/env bash
#go run vendor/github.com/cloudfoundry/cnb-tools/install_tools/main.go "$@"

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

go run -mod=vendor "$DIR/cmd_router.go" package "$@"
