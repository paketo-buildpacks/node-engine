#!/usr/bin/env bash
set -euo pipefail

cd "$( dirname "${BASH_SOURCE[0]}" )/.."

# Install Pack CLI
export PACKBIN=$PWD/.bin
export PATH=$PACKBIN:$PATH
host=$([ $(uname -s) == 'Darwin' ] &&  printf "macos" || printf "linux")
version=$(curl --silent "https://api.github.com/repos/buildpack/pack/releases/latest" | jq -r .tag_name)
wget "https://github.com/buildpack/pack/releases/download/$version/pack-$host.tar.gz" -O $PACKBIN/pack && chmod +x $PACKBIN/pack

# Pull CNB images
export CNB_BUILD_IMAGE=${CNB_BUILD_IMAGE:-cfbuildpacks/cflinuxfs3-cnb-experimental:build}

# TODO: change default to `cfbuildpacks/cflinuxfs3-cnb-experimental:run` when pack cli can use it
export CNB_RUN_IMAGE=${CNB_RUN_IMAGE:-packs/run}

# Always pull latest images
# Most helpful for local testing consistency with CI (which would already pull the latest)
docker pull $CNB_BUILD_IMAGE
docker pull $CNB_RUN_IMAGE

#cd integration

echo "Run Buildpack Runtime Integration Tests"
go test ./integration/... -v -run Integration
