#!/bin/sh -xe

echo "run bootstrap"
GOPATH=/root/go
NECO_DIR=${GOPATH}/src/github.com/cybozu-go/neco
NECO_APPS_DIR=${GOPATH}/src/github.com/cybozu-go/neco-apps
git clone -b ${NECO_BRANCH:-release} https://github.com/cybozu-go/neco.git ${NECO_DIR}
git clone -b ${NECO_APPS_BRANCH:-release} https://github.com/cybozu-go/neco-apps.git ${NECO_APPS_DIR}
git -C ${NECO_DIR} checkout ${NECO_BRANCH:-release}
make -C ${NECO_DIR} clean
make -C ${NECO_DIR}/dctest setup SUDO=""
make -C ${NECO_DIR}/dctest run-placemat-inside-container MENU_ARG=menu-ss.yml SUDO=""
make -C ${NECO_DIR}/dctest test SUITE=bootstrap SUDO=""
git -C ${NECO_APPS_DIR} checkout ${NECO_APPS_BRANCH:-release}
make -C ${NECO_APPS_DIR}/test setup SUDO=""
make -C ${NECO_APPS_DIR}/test dctest SUITE=bootstrap OVERLAY=neco-dev
