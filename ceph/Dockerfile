# Ceph container image

# Stage1: setup environment & install
FROM ghcr.io/cybozu/ubuntu:24.04 AS build

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
ENV DEBIAN_FRONTEND=noninteractive

COPY workspace/*.deb /tmp/
COPY workspace/rocksdb/* /usr/bin/

RUN mkdir -p /var/run/ceph && \
    apt-get update && apt-get install --no-install-recommends -y \
        libstdc++-13-dev jq kmod lvm2 gdisk ca-certificates e2fsprogs attr udev libgflags2.2 \
        /tmp/*.deb && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* && \
    rm -rf /tmp/*.deb && \
    # These settings are commented out in lvm.conf on Ubuntu 24.04, so uncomment them, too.
    sed -i -r \
        -e 's/^(\s*)#?\s*udev_rules = [0-9]+$/\1udev_rules = 0/' \
        -e 's/^(\s*)#?\s*udev_sync = [0-9]+$/\1udev_sync = 0/' \
        -e 's/^(\s*)#?\s*obtain_device_list_from_udev = [0-9]+$/\1obtain_device_list_from_udev = 0/' \
        /etc/lvm/lvm.conf && \
    # validate the sed command worked as expected by checking the effective values
    test "$(lvmconfig --type full activation/udev_sync)" = "udev_sync=0" && \
    test "$(lvmconfig --type full activation/udev_rules)" = "udev_rules=0" && \
    test "$(lvmconfig --type full devices/obtain_device_list_from_udev)" = "obtain_device_list_from_udev=0" && \
    # Clean common files like /tmp, /var/lib, etc.
    rm -rf \
        /etc/{selinux,systemd,udev} \
        /lib/{lsb,udev} \
        /tmp/* \
        /usr/lib{,64}/{locale,systemd,udev,dracut} \
        /usr/share/{doc,info,locale,man} \
        /usr/share/{bash-completion,pkgconfig/bash-completion.pc} \
        /var/log/* \
        /var/tmp/* && \
    find / -xdev \( -name "*.pyc" -o -name "*.pyo" \) -exec rm -f {} \; && \
    mkdir -p /usr/local/share/doc/ceph

COPY workspace/COPYING* /usr/local/share/doc/ceph/

# Stage2: runtime container with squashing layers

FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build / /

RUN chown ceph:ceph -R /run/ceph /var/lib/ceph
