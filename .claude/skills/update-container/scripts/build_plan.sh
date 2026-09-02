#!/bin/bash -e
# build_plan.sh: print the CI build plan for one container directory.
#
# Usage: build_plan.sh DIR
#   DIR : container directory, e.g. "alertmanager" or "golang-all/golang-1.26-noble"
#
# This reproduces what generate_matrix does for a single directory, without
# needing a Pull Request or registry access:
#   - injects the "dir" field into the directory's build-targets.yaml
#   - runs ./build_decision with SKIP_TAG_CHECK=true (skips the ghcr tag lookup)
#
# build_decision validates TAG/BRANCH consistency (./tag_branch_consistency) and
# generates the tag list (./tags_gen) along the way, so a non-zero exit here
# means CI would fail too.
#
# Prints one plan JSON object per line (a directory may declare several images).

DIR="${1%/}"
if [ -z "${DIR}" ]; then
    echo "Usage: build_plan.sh DIR" >&2
    exit 1
fi

cd "$(git rev-parse --show-toplevel)"

BT="${DIR}/build-targets.yaml"
if [ ! -f "${BT}" ]; then
    echo "Error: '${BT}' not found." >&2
    echo "       ceph/ and envoy/ are built by .github/actions/build_{ceph,envoy} instead." >&2
    exit 1
fi

MATRIX=$(mktemp)
trap 'rm -f "${MATRIX}"' EXIT
yq ".[] |= (.dir = \"./${DIR}\")" "${BT}" > "${MATRIX}"

yq -r '.[]."container-image"' "${MATRIX}" | while read -r IMAGE; do
    SKIP_TAG_CHECK=true ./build_decision "${MATRIX}" "${IMAGE}" "./${DIR}"
done
