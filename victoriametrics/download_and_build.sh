#!/bin/bash
set -eo pipefail
curl -fsSL -o victoriametrics.tar.gz "https://github.com/${VICTORIAMETRICS_SRCREPO}/archive/v${VICTORIAMETRICS_VERSION}.tar.gz"
tar -x -z --strip-components 1 -f victoriametrics.tar.gz
rm -f victoriametrics.tar.gz

for P in /*.patch; do
    if [ -f "$P" ]; then
        patch -p1 < $P
    fi
done

BUILDINFO_TAG=v${VICTORIAMETRICS_VERSION} PKG_TAG=v${VICTORIAMETRICS_VERSION} make "$@"
