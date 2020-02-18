#!/usr/bin/env bash
set -eu
set -o pipefail

# shellcheck source=./print.sh
source "$(dirname "${BASH_SOURCE[0]}")/print.sh"

# shellcheck source=./git.sh
source "$(dirname "${BASH_SOURCE[0]}")/git.sh"

function util::tools::path::export() {
    local dir
    dir="${1}"

    if ! echo "${PATH}" | grep -q "${dir}"; then
        PATH="${dir}:$PATH"
        export PATH
    fi
}

function util::tools::pack::install() {
    local dir version os

    util::print::title "Installing pack"

    while [[ "${#}" != 0 ]]; do
      case "${1}" in
        --directory)
          dir="${2}"
          shift 2
          ;;

        --version)
          version="${2}"
          shift 2
          ;;

        *)
          util::print::error "unknown argument \"${1}\""
      esac
    done

    mkdir -p "${dir}"
    util::tools::path::export "${dir}"

    os="$(uname -s)"

    if [[ "${os}" == "Darwin" ]]; then
        os="macos"
    elif [[ "${os}" == "Linux" ]]; then
        os="linux"
    else
        util::print::error "Unsupported operating system"
    fi

    if [[ ! -f "${dir}/pack" ]]; then
        util::print::info "--> installing..."
    elif [[ "$("${dir}/pack" version | cut -d ' ' -f 1)" != *${version}* ]]; then
        rm "${dir}/pack"
        util::print::info "--> updating..."
    else
        util::print::info "--> skipping..."
        return 0
    fi

    GIT_TOKEN="$(util::git::token::fetch)"

    if [[ "${version}" == "latest" ]]; then
        local url
        if [[ "${os}" == "macos" ]]; then
            url="$(
                curl -s \
                    -H "Authorization: token ${GIT_TOKEN}" \
                    https://api.github.com/repos/buildpacks/pack/releases/latest \
                    | jq --raw-output '.assets[1] | .browser_download_url'
                )"
        else
            url="$(
                curl -s \
                    -H "Authorization: token ${GIT_TOKEN}" \
                    https://api.github.com/repos/buildpacks/pack/releases/latest \
                    | jq --raw-output '.assets[0] | .browser_download_url'
                )"
        fi
        util::tools::pack::expand "${dir}" "${url}"
    else
        local tarball
        tarball="pack-${version}-${os}.tar.gz"
        url="https://github.com/buildpacks/pack/releases/download/v${version}/${tarball}"
        util::tools::pack::expand "${dir}" "${url}"
    fi
}

function util::tools::pack::expand() {
    local dir url version
    dir="${1}"
    url="${2}"
    tarball="$(echo "${url}" | sed "s/.*\///")"
    version="v$(echo "${url}" | sed 's/pack-//' | sed 's/-.*//')"

    wget -q "${url}"
    tar xzf "${tarball}" -C "${dir}"
    rm "${tarball}"
}

function util::tools::jam::install () {
    local dir

    while [[ "${#}" != 0 ]]; do
      case "${1}" in
        --directory)
          dir="${2}"
          shift 2
          ;;

        *)
          util::print::error "unknown argument \"${1}\""
      esac
    done

    mkdir -p "${dir}"
    util::tools::path::export "${dir}"

    if [[ ! -f "${dir}/jam" ]]; then
        util::print::title "Installing jam"
        go get -u github.com/cloudfoundry/packit/cargo/jam && \
            go build -o "${dir}/jam" github.com/cloudfoundry/packit/cargo/jam
    fi
}

function util::tools::packager::install () {
    local dir

    while [[ "${#}" != 0 ]]; do
      case "${1}" in
        --directory)
          dir="${2}"
          shift 2
          ;;

        *)
          util::print::error "unknown argument \"${1}\""
      esac
    done

    mkdir -p "${dir}"
    util::tools::path::export "${dir}"

    if [[ ! -f "${dir}/packager" ]]; then
        util::print::title "Installing packager"
        go build -o "${dir}/packager" github.com/cloudfoundry/libcfbuildpack/packager
    fi
}
