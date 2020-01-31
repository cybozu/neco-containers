[![Docker Repository on Quay](https://quay.io/repository/cybozu/ceph/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/ceph)

Ceph container
==============

Build Docker container image for [Ceph][], a distributed object, block, and file storage platform.

Usage
-----

This container image assumes to be used by Rook.
To use in Rook, you need to write a manifest of the custom resource CephCluster with this image.

Reference for updating image
----------------------------

This image based on the Dockerfile produced by `ceph/ceph-container`.
You can make the Dockerfile with following commands:
```console
git clone git@github.com/ceph/ceph-container.git
cd ceph-container
make FLAVORS=luminous,centos,7 stage
# Then the Dockerfile will be located at /staging/luminous-centos-7-x86_64/daemon-base
```

[Ceph]: https://github.com/ceph/ceph
