#!/usr/bin/env bash
set -uo pipefail

cd "$( dirname "${BASH_SOURCE[0]}" )/.."

echo "Run Buildpack Unit Tests"
go test -mod=vendor ./... -v -run Unit
exit_code=$?

if [ "$exit_code" != "0" ]; then
    echo -e "\n\033[0;31m** GO Test Failed **\033[0m"
else
    echo -e "\n\033[0;32m** GO Test Succeeded **\033[0m"
fi

exit $exit_code
