#!/usr/bin/env bash
set -eu
set -o pipefail

readonly PROGDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly BUILDPACKDIR="$(cd "${PROGDIR}/.." && pwd)"

# shellcheck source=.util/tools.sh
source "${PROGDIR}/.util/tools.sh"

# shellcheck source=.util/print.sh
source "${PROGDIR}/.util/print.sh"

# shellcheck source=.util/git.sh
source "${PROGDIR}/.util/git.sh"

function main() {
    if [[ ! -d "${BUILDPACKDIR}/integration" ]]; then
        util::print::warn "** WARNING  No Integration tests **"
    fi

    tools::install
    images::pull
    token::fetch
    tests::run
}

function tools::install() {
    util::tools::pack::install \
        --directory "${BUILDPACKDIR}/.bin" \
        --version "latest"

    if [[ -f "${BUILDPACKDIR}/.packit" ]]; then
        util::tools::jam::install \
            --directory "${BUILDPACKDIR}/.bin"

    else
        util::tools::packager::install \
            --directory "${BUILDPACKDIR}/.bin"
    fi
}

function images::pull() {
    util::print::title "Pulling build image..."
    docker pull "${CNB_BUILD_IMAGE:=cloudfoundry/build:full-cnb}"

    util::print::title "Pulling run image..."
    docker pull "${CNB_RUN_IMAGE:=cloudfoundry/run:full-cnb}"

    util::print::title "Pulling cflinuxfs3 builder image..."
    docker pull "${CNB_BUILDER_IMAGE:=cloudfoundry/cnb:cflinuxfs3}"

    export CNB_BUILD_IMAGE
    export CNB_RUN_IMAGE
    export CNB_BUILDER_IMAGE

    util::print::title "Setting default pack builder image..."
    pack set-default-builder "${CNB_BUILDER_IMAGE}"
}

function token::fetch() {
    GIT_TOKEN="$(util::git::token::fetch)"
    export GIT_TOKEN
}

function tests::run() {
    util::print::title "Run Buildpack Runtime Integration Tests"
    pushd "${BUILDPACKDIR}" > /dev/null
        if GOMAXPROCS=4 go test -timeout 0 ./integration/... -v -mod=vendor -run Integration; then
            util::print::success "** GO Test Succeeded **"
        else
            util::print::error "** GO Test Failed **"
        fi
    popd > /dev/null
}

main "${@:-}"
