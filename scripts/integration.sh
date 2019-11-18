#!/usr/bin/env bash
set -euo pipefail

cd "$( dirname "${BASH_SOURCE[0]}" )/.."

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

if [[ ! -d integration ]]; then
    echo -e "\n\033[0;31m** WARNING  No Integration tests **\033[0m"
    exit 0
fi

PACK_VERSION=${PACK_VERSION:-"latest"}
source "$SCRIPT_DIR"/install_tools.sh "$PACK_VERSION"

export CNB_BUILD_IMAGE=${CNB_BUILD_IMAGE:-cloudfoundry/build:full-cnb}
export CNB_RUN_IMAGE=${CNB_RUN_IMAGE:-cloudfoundry/run:full-cnb}

# Always pull latest images
# Most helpful for local testing consistency with CI (which would already pull the latest)
docker pull "$CNB_BUILD_IMAGE"
docker pull "$CNB_RUN_IMAGE"

# Get GIT_TOKEN for github rate limiting
GIT_TOKEN=${GIT_TOKEN:-"$(lpass show Shared-CF\ Buildpacks/concourse-private.yml | grep buildpacks-github-token | cut -d ' ' -f 2)"}
export GIT_TOKEN

echo "Run Buildpack Runtime Integration Tests"
set +e
GOMAXPROCS=4 go test -timeout 0 ./integration/... -v -mod=vendor -run Integration
exit_code=$?

if [[ "$exit_code" != "0" ]]; then
    echo -e "\n\033[0;31m** GO Test Failed **\033[0m"
else
    echo -e "\n\033[0;32m** GO Test Succeeded **\033[0m"
fi

exit "$exit_code"
