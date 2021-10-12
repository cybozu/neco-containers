#!/bin/sh
set -eu

CEPH_DIR=$(readlink -f $(dirname $0))

if [ $# -ne 1 ]; then
    echo "deb.sh VERSION"
    exit 1
fi

VERSION="$1"

# Checkout Ceph source
mkdir -p src/workspace
cd src
git clone -b v${VERSION} --depth=1 --recurse-submodules --shallow-submodules https://github.com/ceph/ceph.git
cd ceph

# Install dependencies
apt-get update
./install-deps.sh

# Build Ceph packages
sed -i -e 's/WITH_CEPHFS_JAVA=ON/WITH_CEPHFS_JAVA=OFF/' debian/rules
sed -i -e 's@usr/share/java/libcephfs-test.jar@@' debian/ceph-test.install
rm debian/libcephfs-java.jlibs debian/libcephfs-jni.install debian/ceph-mgr-dashboard*
# To avoid OOM killer, use 10 parallelism instead of 20 (max vCPU).
dpkg-buildpackage --build=binary -uc -us -j10
rm ../*-dbg_*.deb ../*-dev_*.deb
mv ../*.deb ../workspace/
mv COPYING* ../workspace
