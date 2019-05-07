#!/usr/bin/env bash
set -euo pipefail

cd "$( dirname "${BASH_SOURCE[0]}" )/.."

target_os=${1:-linux}

for b in $(ls cmd); do
    GOOS="$target_os" go build -mod=vendor -ldflags="-s -w" -o "bin/$b" "cmd/$b/main.go"
done
