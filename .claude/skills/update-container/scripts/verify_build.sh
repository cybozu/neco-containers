#!/bin/bash
# verify_build.sh: build one container directory exactly the way CI does.
#
# Usage: verify_build.sh DIR
#   DIR : container directory, e.g. "alertmanager"
#
# Mirrors .github/actions/build_push/action.yaml, for each plan build_plan.sh
# emits:
#   1. make -C DIR <make-targets>
#   2. docker buildx build --platform P [--target T] [--load] -t <tags...> DIR
#   3. IMAGE_TAG=<primary tag> make -C DIR <make-post-targets>
#
# Nothing is pushed. The build context is DIR and the Dockerfile is
# DIR/Dockerfile; CI passes no build-args, so neither do we.
set -euo pipefail

DIR="${1:?Usage: verify_build.sh DIR}"
DIR="${DIR%/}"
HERE="$(cd "$(dirname "$0")" && pwd)"

if ! command -v docker > /dev/null; then
    echo "Error: docker with buildx is required; this script reproduces the CI build locally." >&2
    exit 1
fi

cd "$(git rev-parse --show-toplevel)"

# One jq field per line, so a whole plan is read with one readarray. jq also
# assembles the buildx flags, which keeps conditional flag juggling out of bash.
FIELDS='."container-image", .tags[0], (.platforms // "linux/amd64"),
        (."make-targets" // ""), (."make-post-targets" // "")'
FLAGS='"--platform", (.platforms // "linux/amd64"),
       (select(.target)  | "--target", .target),
       (select(.load)    | "--load"),
       (.tags[]          | "-t", .)'

# Echo the command, then run it.
run() { echo "--> $*"; "$@"; }

# Plans arrive on fd 3 so that make/docker cannot swallow them from stdin.
while IFS= read -r plan <&3; do
    [ -n "${plan}" ] || continue
    readarray -t field < <(jq -r "${FIELDS}" <<< "${plan}")
    readarray -t flags < <(jq -r "${FLAGS}"  <<< "${plan}")
    image=${field[0]} primary=${field[1]} platforms=${field[2]}
    pre=${field[3]} post=${field[4]}

    echo "==> ${image}: building ${primary} (platforms=${platforms})"
    case "${platforms}" in
        *,*) echo "Note: multi-platform needs 'docker buildx create --use' and QEMU." >&2 ;;
    esac

    for t in ${pre}; do
        run make -C "${DIR}" "${t}"
    done

    run docker buildx build "${flags[@]}" "${DIR}"

    for t in ${post}; do
        run env IMAGE_TAG="${primary}" make -C "${DIR}" "${t}"
    done

    echo "==> ${image}: OK (${primary})"
done 3< <("${HERE}/build_plan.sh" "${DIR}")
