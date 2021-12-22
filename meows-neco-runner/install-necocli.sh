#!/bin/bash

BIN_DIR=${INSTALL_DIR-/tmp/neco-operation-cli/bin}
TMP_DIR=/tmp

mkdir -p ${BIN_DIR} ${TMP_DIR}

curl -sSL -o ${BIN_DIR}/jq https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64
chmod +x ${BIN_DIR}/jq

DATE=$(curl -sSL "https://api.github.com/repos/cybozu-go/neco/releases/latest" | ${BIN_DIR}/jq -r ".tag_name" | sed -e "s/release-//")
FILE="neco-operation-cli-linux_${DATE}_amd64.deb"
curl -sSL -o ${TMP_DIR}/${FILE} https://github.com/cybozu-go/neco/releases/download/release-${DATE}/${FILE}
dpkg -x ${TMP_DIR}/${FILE} ${TMP_DIR}
mv ${TMP_DIR}/usr/bin/* ${BIN_DIR}

echo ${BIN_DIR} >> $GITHUB_PATH
