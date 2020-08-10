#!/usr/bin/env bash
set -eu
set -o pipefail

readonly PROGDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly BUILDPACKDIR="$(cd "${PROGDIR}/.." && pwd)"

# shellcheck source=SCRIPTDIR/.util/print.sh
source "${PROGDIR}/.util/print.sh"

function main() {
    util::print::title "Run Buildpack Unit Tests"

    pushd "${BUILDPACKDIR}" > /dev/null
        if go test ./... -v -run Unit; then
            util::print::success "** GO Test Succeeded **"
        else
            util::print::error "** GO Test Failed **"
        fi
    popd > /dev/null
}

main "${@:-}"
