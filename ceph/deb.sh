#!/bin/sh
set -eu

CEPH_DIR=$(readlink -f $(dirname $0))

print_usage() {
    echo "Usage: deb.sh [-a] VERSION"
}

# Process optional arguments
WITH_ASAN=OFF
while getopts a OPT
do
    case $OPT in
        a) WITH_ASAN=ON;;
        \?) print_usage; exit 1;;
    esac
done

# Process mandatory arguments
shift $((${OPTIND}-1))
if [ $# -ne 1 ]; then
    print_usage
    exit 1
fi

VERSION="$1"

# Checkout Ceph source
mkdir -p src/workspace/dev/
cd src
git clone -b v${VERSION} --depth=1 --recurse-submodules --shallow-submodules https://github.com/ceph/ceph.git
cd ceph

# Apply temporary patch
git apply ${CEPH_DIR}/43581.patch
git apply ${CEPH_DIR}/44413.patch
git apply ${CEPH_DIR}/fix_pytest_version.patch
if [ "$WITH_ASAN" = "ON" ]; then
    echo "WITH_ASAN is ON. ASAN patch will be applied."
    git apply ${CEPH_DIR}/asan.patch
fi

# Install dependencies
apt-get update
./install-deps.sh

# Build Ceph packages
sed -i -e 's/WITH_CEPHFS_JAVA=ON/WITH_CEPHFS_JAVA=OFF/' debian/rules
sed -i -e 's@usr/share/java/libcephfs-test.jar@@' debian/ceph-test.install
rm debian/libcephfs-java.jlibs debian/libcephfs-jni.install debian/ceph-mgr-dashboard*
# To avoid OOM killer, use 10 parallelism instead of 20 (max vCPU).
dpkg-buildpackage --build=binary -uc -us -j10
rm ../*-dbg_*.deb
mv ../*-dev_*.deb ../workspace/dev/
mv ../*.deb ../workspace/
mv COPYING* ../workspace
