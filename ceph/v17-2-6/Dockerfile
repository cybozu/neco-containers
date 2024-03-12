# Ceph container image

# Stage1: setup environment & install
FROM quay.io/cybozu/ubuntu:20.04 AS build

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
ENV DEBIAN_FRONTEND noninteractive

COPY workspace/*.deb /tmp/
COPY workspace/rocksdb/* /usr/bin/

RUN mkdir -p /var/run/ceph && \
    apt-get update && apt-get install --no-install-recommends -y \
        libstdc++-9-dev jq kmod lvm2 gdisk ca-certificates e2fsprogs attr udev libgflags2.2 \
        /tmp/*.deb && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* && \
    rm -rf /tmp/*.deb && \
    sed -i -e 's/udev_rules = 1/udev_rules = 0/' -e 's/udev_sync = 1/udev_sync = 0/' -e 's/obtain_device_list_from_udev = 1/obtain_device_list_from_udev = 0/' /etc/lvm/lvm.conf && \
    # validate the sed command worked as expected
    grep -sqo "udev_sync = 0" /etc/lvm/lvm.conf && \
    grep -sqo "udev_rules = 0" /etc/lvm/lvm.conf && \
    grep -sqo "obtain_device_list_from_udev = 0" /etc/lvm/lvm.conf && \
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

COPY --from=build / /

RUN chown ceph:ceph -R /run/ceph /var/lib/ceph
