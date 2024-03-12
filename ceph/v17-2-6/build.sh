#!/bin/sh
set -eu

CEPH_DIR=$(readlink -f $(dirname $0))

if [ $# -ne 1 ]; then
    echo "Usage: $0 VERSION"
    exit 1
fi

VERSION="$1"

# Checkout Ceph source
mkdir -p src/workspace/rocksdb/
cd src
git clone -b v${VERSION} --depth=1 --recurse-submodules --shallow-submodules https://github.com/ceph/ceph.git
cd ceph

# Install dependencies
sudo apt-get update
./install-deps.sh

# Prebuild ceph source to generate files in `src/pybind/mgr/dashboard/frontend/dist` needed by CMake
./make-dist

# Build Ceph packages
sed -i -e 's/WITH_CEPHFS_JAVA=ON/WITH_CEPHFS_JAVA=OFF/' debian/rules
sed -i -e 's@usr/share/java/libcephfs-test.jar@@' debian/ceph-test.install
rm debian/libcephfs-java.jlibs debian/libcephfs-jni.install debian/ceph-mgr-dashboard*
# To avoid OOM killer, use 10 parallelism instead of 20 (max vCPU).
dpkg-buildpackage --build=binary -uc -us -j10
rm ../*-dbg_*.deb ../ceph-test_*.deb
mv ../*.deb ../workspace/
mv COPYING* ../workspace

# Intall libgflags to build rocksdb tools
sudo apt-get install --no-install-recommends -y libgflags-dev
# Build rocksdb tools
make -C src/rocksdb release -j10
find src/rocksdb -maxdepth 1 -type f -executable -exec mv {} ../workspace/rocksdb \;
