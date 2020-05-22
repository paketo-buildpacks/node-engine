#!/usr/bin/env bash
set -eu
set -o pipefail

readonly PROGDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly BUILDPACKDIR="$(cd "${PROGDIR}/.." && pwd)"

function main() {
    mkdir -p "${BUILDPACKDIR}/bin"

    if [[ -f "${BUILDPACKDIR}/main.go" ]]; then
        pushd "${BUILDPACKDIR}/bin" > /dev/null || return
            printf "%s" "Building run..."

            GOOS=linux \
              go build \
                -ldflags="-s -w" \
                -o "run" \
                  "${BUILDPACKDIR}"

            echo "Success!"

            for name in detect build; do
              printf "%s" "Linking ${name}..."

              ln -sf "run" "${name}"

              echo "Success!"
            done
        popd > /dev/null || return
    fi

    if [[ -d "${BUILDPACKDIR}/cmd" ]]; then
        local name
        for src in "${BUILDPACKDIR}"/cmd/*; do
            name="$(basename "${src}")"

            printf "%s" "Building ${name}..."

            GOOS="linux" \
                go build \
                    -ldflags="-s -w" \
                    -o "${BUILDPACKDIR}/bin/${name}" \
                        "${src}/main.go"

            echo "Success!"
        done
    fi
}

main "${@:-}"
