Go container
============

This directory provides a Dockerfile to build a Docker container
that includes [Go](https://golang.org/) language runtime and following
tools:

* [goimports](https://godoc.org/golang.org/x/tools/cmd/goimports)
* [golint](https://github.com/golang/lint)
* [ineffassign](https://github.com/gordonklaus/ineffassign)
* [ghr](https://github.com/tcnksm/ghr)
* [custom-checker](./analyzer/cmd/custom-checker/README.md)
* [eventuallycheck](./analyzer/cmd/eventuallycheck/README.md)
* [restrictpkg](./analyzer/cmd/restrictpkg/README.md)

This container is based on [quay.io/cybozu/ubuntu-dev](https://quay.io/repository/cybozu/ubuntu-dev).

Docker images
-------------

Docker images are available on [Quay.io](https://quay.io/repository/cybozu/golang)
