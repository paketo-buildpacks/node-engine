#!/usr/bin/env bash
set -eu
set -o pipefail

readonly PROGDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly BUILDPACKDIR="$(cd "${PROGDIR}/.." && pwd)"

function main() {
    local name
    for src in "${BUILDPACKDIR}"/cmd/*; do
        name="$(basename "${src}")"

        printf "%s" "Building ${name}..."

        GOOS="linux" \
            go build \
                -mod=vendor \
                -ldflags="-s -w" \
                -o "${BUILDPACKDIR}/bin/${name}" \
                    "${src}/main.go"

        echo "Success!"
    done
}

main "${@:-}"
