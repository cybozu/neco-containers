cert-manager container
======================

This directory provides a Dockerfile to build a Docker container
that runs [cert-manager](https://github.com/cert-manager/cert-manager).

Usage
-----

### Start `cert-manager`

Run the container

```console
$ docker run -d --read-only --name=cert-manager-controller \
    ghcr.io/cybozu/cert-manager:1.18 controller
```

License
-------

[LICENSES](https://github.com/cert-manager/cert-manager/blob/master/LICENSES) is a file bundled with all LICENSE files under the `vendor` directory.

Docker images
-------------

Docker images are available on [ghcr.io](https://github.com/cybozu/neco-containers/pkgs/container/cert-manager)
