#!/bin/bash -e
# Update vendored pomerium source
if [ "$(basename "${PWD}")" != "pomerium" ]; then
    echo "pomerium/update-src.sh must be run in pomerium directory" >&2
    exit 1
fi

SRC_DIR=src

POMERIUM_VERSION=$(sed -n 's/^ARG POMERIUM_VERSION=//p' Dockerfile)
POMERIUM_CHECKSUM=$(sed -n 's/^ARG POMERIUM_CHECKSUM=//p' Dockerfile)

curl -fsSL -o "./pomerium.tar.gz" "https://github.com/pomerium/pomerium/archive/v${POMERIUM_VERSION}.tar.gz"
echo "${POMERIUM_CHECKSUM}  ./pomerium.tar.gz" | sha256sum -c -

rm -rf "${SRC_DIR}"
mkdir "${SRC_DIR}"
tar zxf "./pomerium.tar.gz" -C "${SRC_DIR}" --strip-components 1
rm -f ./pomerium.tar.gz
