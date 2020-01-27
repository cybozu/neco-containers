# Ceph container image

# Stage1: build from source
FROM quay.io/cybozu/ubuntu-dev:18.04 AS build

ARG CEPH_VERSION=14.2.6

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN apt-get update && apt-get install --no-install-recommends -y git
RUN git clone https://github.com/ceph/ceph.git /work/ceph

WORKDIR /work/ceph
RUN git checkout v${CEPH_VERSION} \
    && git submodule update --init --recursive
RUN ./install-deps.sh
RUN sed -i -e 's/WITH_CEPHFS_JAVA=ON/WITH_CEPHFS_JAVA=OFF/' debian/rules
RUN sed -i -e 's/usr\/share\/java\/libcephfs-test.jar//' debian/ceph-test.install
RUN rm debian/libcephfs-java.jlibs debian/libcephfs-jni.install debian/ceph-mgr-dashboard*
RUN dpkg-buildpackage --build=binary -uc -us -j20
RUN rm ../*-dbg_*.deb ../*-dev_*.deb

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:18.04 AS build2

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
ENV DEBIAN_FRONTEND noninteractive

COPY --from=build /work/*.deb /tmp/
RUN apt-get update && apt-get install --no-install-recommends -y \
        jq kmod lvm2 gdisk ca-certificates e2fsprogs attr udev \
        /tmp/*.deb \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* \
    && rm -rf /tmp/*.deb

RUN sed -i -e 's/udev_rules = 1/udev_rules = 0/' -e 's/udev_sync = 1/udev_sync = 0/' -e 's/obtain_device_list_from_udev = 1/obtain_device_list_from_udev = 0/' /etc/lvm/lvm.conf && \
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
    find  / -xdev \(-name "*.pyc"\) -o \(-name "*.pyo"\) -exec rm -f {} \;

RUN ln -s /sbin/sgdisk /usr/sbin/sgdisk

# Stage3: squash layers
FROM scratch

COPY --from=build2 / /
