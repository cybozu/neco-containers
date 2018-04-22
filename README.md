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

License
-------

MIT

[quay]: https://quay.io/organization/cybozu
