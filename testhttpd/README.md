testhttpd container
===============

This directory provides container image `testhttpd` and contains its source code.

testhttpd is a micro HTTP server that can run in Kubernetes cluster with limited privileges.
Specifically, it runs as a non-root user and does not write to the root filesystem.


Usage
-----

```console
$ kubectl run quay.io/cybozu/testhttpd
``` 

Access from some clients like below.

```
$ curl http://<serving address>:8000
```

If you want a delayed response, you can give the delay as a query(`sleep`).

```
$ curl http://<serving address>:8000/?sleep=10s
```

Docker images
-------------

Docker images are available on [Quay.io](https://quay.io/repository/cybozu/testhttpd)
