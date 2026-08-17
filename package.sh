#!/usr/bin/env bash

# Build gocron and gocron-node release archives with the shared Go packager.
# Example: ./package.sh -a amd64 -p 'linux windows' -v v2.0.3

set -o errexit
set -o pipefail

ARGS=()

while getopts 'p:a:v:' option; do
    case "${option}" in
        p)
            ARGS+=("-os" "${OPTARG}")
            ;;
        a)
            ARGS+=("-arch" "${OPTARG}")
            ;;
        v)
            ARGS+=("-version" "${OPTARG}")
            ;;
        *)
            echo 'Usage: package.sh [-p "linux windows"] [-a "386 amd64"] [-v v2.0.3]'
            exit 1
            ;;
    esac
done

exec go run ./tools/package-release "${ARGS[@]}"
