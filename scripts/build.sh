#!/usr/bin/env bash
set -exuo pipefail

cd "$( dirname "${BASH_SOURCE[0]}" )/.."
TARGET_OS=${1:-linux}

for b in cmd/*; do
    echo -n "Building $b..."
    GOOS=$TARGET_OS go build -mod=vendor -ldflags="-s -w" -o bin/$(basename $b) $b/main.go
    echo "done"
done
