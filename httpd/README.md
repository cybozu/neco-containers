[![Docker Repository on Quay](https://quay.io/repository/cybozu/httpd/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/httpd)

httpd container
===============

This directory provides container image `httpd` and contains its source code.

`httpd` is a simple HTTP server for testing (*not for production*)
in Kubernetes cluster which is enabled PodSecurityPolicy admission plugin.


Usage
-----

```console
$ kubectl run quay.io/cybozu/httpd
``` 
