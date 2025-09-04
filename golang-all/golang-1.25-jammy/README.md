Go container
============

This directory provides a Dockerfile to build a Docker container
that includes [Go](https://golang.org/) language runtime and following
tools:

* [goimports](https://godoc.org/golang.org/x/tools/cmd/goimports)
* [golint](https://github.com/golang/lint)
* [staticcheck](https://staticcheck.io/)
* [ineffassign](https://github.com/gordonklaus/ineffassign)
* [ghr](https://github.com/tcnksm/ghr)
* [golang custom analyzer](https://github.com/cybozu-go/golang-custom-analyzer)

This container is based on [ghcr.io/cybozu/ubuntu-dev](https://ghcr.io/repository/cybozu/ubuntu-dev).

Docker images
-------------

Docker images are available on [ghcr.io](https://github.com/cybozu/neco-containers/pkgs/container/golang)
