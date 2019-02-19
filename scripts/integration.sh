#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

export PATH="$DIR/../.bin":$PATH
go run -mod=vendor "$DIR/cmd_router.go" integration "$@"
