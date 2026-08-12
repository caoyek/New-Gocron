#!/usr/bin/env bash

# Build gocron and gocron-node release archives.
# Example: ./package.sh -a amd64 -p linux -v v2.0.0

set -o errexit
set -o nounset
set -o pipefail

BINARY_NAME=''
MAIN_PACKAGE=''
VERSION=''
GIT_COMMIT_ID=''
INPUT_OS=()
INPUT_ARCH=()
DEFAULT_OS=$(go env GOHOSTOS)
DEFAULT_ARCH=$(go env GOHOSTARCH)
SUPPORT_OS=(linux darwin windows)
SUPPORT_ARCH=(386 amd64)
LDFLAGS=''
INCLUDE_FILES=(README.md LICENSE)
PACKAGE_DIR=''
BUILD_DIR=''

print_message_and_exit() {
    if [[ -n "${1}" ]]; then
        echo "${1}"
    fi
    exit 1
}

is_supported() {
    local expected="${1}"
    shift
    local value=''
    for value in "$@"; do
        if [[ "${value}" == "${expected}" ]]; then
            return 0
        fi
    done
    return 1
}

set_os_arch() {
    if [[ ${#INPUT_OS[@]} -eq 0 ]]; then
        INPUT_OS=("${DEFAULT_OS}")
    fi
    if [[ ${#INPUT_ARCH[@]} -eq 0 ]]; then
        INPUT_ARCH=("${DEFAULT_ARCH}")
    fi

    local os=''
    local arch=''
    for os in "${INPUT_OS[@]}"; do
        is_supported "${os}" "${SUPPORT_OS[@]}" || print_message_and_exit "Unsupported operating system: ${os}"
    done
    for arch in "${INPUT_ARCH[@]}"; do
        is_supported "${arch}" "${SUPPORT_ARCH[@]}" || print_message_and_exit "Unsupported CPU architecture: ${arch}"
    done
}

resolve_version() {
    if [[ -n "${VERSION}" ]]; then
        return
    fi
    VERSION=$(git describe --tags --abbrev=0 2>/dev/null || true)
    if [[ -z "${VERSION}" ]]; then
        VERSION='v2.0.0'
    fi
}

init_package() {
    set_os_arch
    resolve_version

    GIT_COMMIT_ID=$(git rev-parse --short HEAD)
    local build_date=''
    build_date=$(date -u '+%Y-%m-%dT%H:%M:%SZ')
    LDFLAGS="-w -X main.AppVersion=${VERSION} -X main.BuildDate=${build_date} -X main.GitCommit=${GIT_COMMIT_ID}"
    PACKAGE_DIR="${BINARY_NAME}-package"
    BUILD_DIR="${BINARY_NAME}-build"

    rm -rf "${BUILD_DIR}" "${PACKAGE_DIR}"
    mkdir -p "${BUILD_DIR}" "${PACKAGE_DIR}"
}

build_binary() {
    local os=''
    local arch=''
    local filename=''
    local target_dir=''
    for os in "${INPUT_OS[@]}"; do
        for arch in "${INPUT_ARCH[@]}"; do
            filename="${BINARY_NAME}"
            if [[ "${os}" == 'windows' ]]; then
                filename="${filename}.exe"
            fi
            target_dir="${BUILD_DIR}/${BINARY_NAME}-${os}-${arch}"
            mkdir -p "${target_dir}"
            env CGO_ENABLED=0 GOOS="${os}" GOARCH="${arch}" \
                go build -ldflags "${LDFLAGS}" -o "${target_dir}/${filename}" "${MAIN_PACKAGE}"
        done
    done
}

copy_release_files() {
    local target_dir="${1}"
    local item=''
    for item in "${INCLUDE_FILES[@]}"; do
        cp "${item}" "${target_dir}/"
    done
}

archive_binary() {
    local os=''
    local arch=''
    local target_name=''
    for os in "${INPUT_OS[@]}"; do
        for arch in "${INPUT_ARCH[@]}"; do
            target_name="${BINARY_NAME}-${os}-${arch}"
            copy_release_files "${BUILD_DIR}/${target_name}"
            if [[ "${os}" == 'windows' ]]; then
                (cd "${BUILD_DIR}" && zip -rq "../${PACKAGE_DIR}/${BINARY_NAME}-${VERSION}-${os}-${arch}.zip" "${target_name}")
            else
                tar czf "${PACKAGE_DIR}/${BINARY_NAME}-${VERSION}-${os}-${arch}.tar.gz" -C "${BUILD_DIR}" "${target_name}"
            fi
        done
    done
}

run_package() {
    init_package
    build_binary
    archive_binary
    rm -rf "${BUILD_DIR}"
}

package_gocron() {
    BINARY_NAME='gocron'
    MAIN_PACKAGE='./cmd/gocron'
    run_package
}

package_gocron_node() {
    BINARY_NAME='gocron-node'
    MAIN_PACKAGE='./cmd/node'
    run_package
}

while getopts 'p:a:v:' option; do
    case "${option}" in
        p)
            read -r -a INPUT_OS <<< "${OPTARG//,/ }"
            ;;
        a)
            read -r -a INPUT_ARCH <<< "${OPTARG//,/ }"
            ;;
        v)
            VERSION="${OPTARG}"
            ;;
        *)
            print_message_and_exit 'Usage: package.sh [-p "linux windows"] [-a "386 amd64"] [-v v2.0.0]'
            ;;
    esac
done

package_gocron
package_gocron_node
