#!/usr/bin/env bash
set -eu
set -o pipefail

readonly PROGDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly BUILDPACKDIR="$(cd "${PROGDIR}/.." && pwd)"

# shellcheck source=.util/tools.sh
source "${PWD}/scripts/.util/tools.sh"

# shellcheck source=.util/print.sh
source "${PWD}/scripts/.util/print.sh"

if ! command -v realpath > /dev/null; then
  function realpath() {
      [[ "${1}" = /* ]] && echo "${1}" || echo "${PWD}/${1#./}"
  }
fi

function main() {
    local full_path args version cached archive offline
    PACKAGE_DIR=${PACKAGE_DIR:-"${BUILDPACKDIR}/$(basename ${BUILDPACKDIR})_$(openssl rand -hex 4)"}

    full_path="$(realpath "${PACKAGE_DIR}")"

    while [[ "${#}" != 0 ]]; do
      case "${1}" in
        --archive|-a)
          archive="true"
          shift 1
          ;;

        --cached|-c)
          cached="true"
          offline="true"
          shift 1
          ;;

        --version|-v)
          version="${2}"
          shift 2
          ;;

        "")
          # skip if the argument is empty
          shift 1
          ;;

        *)
          util::print::error "unknown argument \"${1}\""
      esac
    done

    if [[ -f "${BUILDPACKDIR}/.packit" ]]; then
        #use jam
        util::tools::jam::install --directory "${BUILDPACKDIR}/.bin"
        if [[ -z "${version:-}" ]]; then #version not provided, use latest git tag
            git_tag=$(git describe --abbrev=0 --tags)
            version=${git_tag:1}
        fi

        extra_args=""

        if [[ -n "${offline:-}" ]]; then
            PACKAGE_DIR="${PACKAGE_DIR}-cached"
            extra_args+="--offline"
        fi

        if [[ "${PACKAGE_DIR}" != "*.tgz" ]]; then
            PACKAGE_DIR="${PACKAGE_DIR}.tgz"
        fi


        .bin/jam pack \
        --buildpack "$(pwd)/buildpack.toml" \
        --version "${version}" \
        --output "${PACKAGE_DIR}" \
        ${extra_args}

    else
        # use old packager
        util::tools::packager::install --directory "${BUILDPACKDIR}/.bin"

        args="${BUILDPACKDIR}/.bin/packager"
        if [[ -n "${cached:-}" ]]; then
            full_path="${full_path}-cached"
        else
            args="${args} --uncached"
        fi

        if [[ -n "${archive:-}" ]]; then
            args="${args} -archive"
        fi

        if [[ -z "${version:-}" ]]; then
            version="$(cd "${BUILDPACKDIR}" && git describe --abbrev=0 --tags)"
        fi

        args="${args} -version ${version}"

        pushd "${BUILDPACKDIR}" > /dev/null
            eval "${args}" "${full_path}"
        popd > /dev/null

        if [[ -n "${BP_REWRITE_HOST:-}" ]]; then
            sed -i '' -e "s|^uri = \"https:\/\/buildpacks\.cloudfoundry\.org\(.*\)\"$|uri = \"http://${BP_REWRITE_HOST}\1\"|g" "${full_path}/buildpack.toml"
        fi
    fi



}

main "${@:-}"
