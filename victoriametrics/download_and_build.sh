#!/bin/bash
set -eo pipefail
curl -fsSL -o victoriametrics.tar.gz "https://github.com/${VICTORIAMETRICS_SRCREPO}/archive/v${VICTORIAMETRICS_VERSION}.tar.gz" \
    && tar -x -z --strip-components 1 -f victoriametrics.tar.gz \
    && rm -f victoriametrics.tar.gz \
    && BUILDINFO_TAG=v${VICTORIAMETRICS_VERSION} PKG_TAG=v${VICTORIAMETRICS_VERSION} make "$@"
