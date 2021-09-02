#!/bin/sh -xe

echo "setup secret for cloud dns"
cp /secrets/account.json ${NECO_APPS_DIR}/test/

echo "run bootstrap"
git -C ${NECO_DIR} checkout ${NECO_BRANCH:-release}
make -C ${NECO_DIR} clean
make -C ${NECO_DIR}/dctest setup
make -C ${NECO_DIR}/dctest run-placemat-inside-container MENU_ARG=menu-ss.yml
make -C ${NECO_DIR}/dctest test SUITE=bootstrap
git -C ${NECO_APPS_DIR} checkout ${NECO_APPS_BRANCH:-release}
make -C ${NECO_APPS_DIR}/test setup
make -C ${NECO_APPS_DIR}/test dctest SUITE=bootstrap OVERLAY=neco-dev
