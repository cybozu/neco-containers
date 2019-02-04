[![CircleCI](https://circleci.com/gh/cybozu/neco-containers.svg?style=svg)](https://circleci.com/gh/cybozu/neco-containers)

Neco Containers
===============

This repository contains Dockerfiles to build OSS products
used in our project, Neco.  They are built from the official
sources, and based on our Ubuntu base image.

See also: [github.com/cybozu/ubuntu-base](https://github.com/cybozu/ubuntu-base).

Built images can be pulled from [quay.io/cybozu][quay].

How it works
------------

Subdirectories in this repository have `TAG` and `BRANCH` files
in addition to files to build Docker images.

These will be used by CircleCI to tag the built images.
CircleCI does the following each time commits are pushed to a branch.

1. For each directory containing `TAG` file:
    1. Read `TAG` file and check if the repository at [quay.io/cybozu][quay] with the same name of the directory.
    1. If the repository contains the same tag in `TAG`, continue to the next directory.
    1. Otherwise, build a Docker image using `Dockerfile` under the directory.
1. If the branch is not `master`, CircleCI stops here without pushing.
1. If the branch is `master`, for each directory with a built image:
    1. Tag the built image with tag in `TAG` file.
    1. Push the tagged image to quay.io.
    1. If the directory contains `BRANCH` file:
        1. Tag the built image with tag in `BRANCH` file.
        1. Push the tagged image to quay.io.

Tag naming
----------

If the image is built for an upstream version X.Y.Z, the first image tag _must_ be X.Y.Z.1.
The last version indicates the container image version and _must_ be incremented when some
changes are introduced to the image.

If the upstream version has no patch version (X.Y), fill the patch version with 0 then
add the container image version A (X.Y.0.A).

If the upstream version is a Debian package as X.Y.Z-PACKAGE (note that -PACKAGE
should not be confused as pre-release as in semver), use "X.Y.Z.PACKAGE" as the
upstream version and add the container image version as the fifth version.

The container image version _must_ be reset to 1 when the upstream version is changed.

### Example

If the upstream version is "1.2.0-beta.3", the image tag must begin with "1.2.0-beta.3.1".

Branch naming
-------------

If the image is built for an upstream version X.Y.Z, the branch name _should_ be X.Y
for X > 0, or "0" for X == 0.

License
-------

MIT

[quay]: https://quay.io/organization/cybozu
