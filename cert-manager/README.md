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
    quay.io/cybozu/cert-manager:1.7 controller
```

License
-------

[LICENSES](https://github.com/cert-manager/cert-manager/blob/master/LICENSES) is a file bundled with all LICENSE files under the `vendor` directory.

Docker images
-------------

Docker images are available on [Quay.io](https://quay.io/repository/cybozu/cert-manager)
