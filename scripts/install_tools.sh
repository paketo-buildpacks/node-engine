#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

go run -mod=vendor "$DIR/cmd_router.go" install_tools "$@"
