[![Docker Repository on Quay](https://quay.io/repository/cybozu/golang/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/golang)

Go container
============

This directory provides a Dockerfile to build a Docker container
that includes [Go](https://golang.org/) language runtime and following
tools:

* [goimports](https://godoc.org/golang.org/x/tools/cmd/goimports)
* [golint](https://github.com/golang/lint)
* [ineffassign](https://github.com/gordonklaus/ineffassign)
* [ghr](https://github.com/tcnksm/ghr)
* [nilerr](https://github.com/gostaticanalysis/nilerr)
* [custome-checker](./analyzer/cmd/custome-checker/README.md)
* [eventuallycheck](./analyzer/cmd/eventuallycheck/README.md)
* [restrictpkg](./analyzer/cmd/restrictpkg/README.md)

This container is based on [quay.io/cybozu/ubuntu-dev](https://quay.io/repository/cybozu/ubuntu-dev).
