[![GoDoc](https://godoc.org/github.com/cybozu/neco-containers/cke-tools/src?status.svg)][godoc]
[![Go Report Card](https://goreportcard.com/badge/github.com/cybozu/neco-containers/cke-tools/src)](https://goreportcard.com/report/github.com/cybozu-go/cke-tools)

cke-tools
=========

CKE tools is a suite of the various tools used by [CKE][].
This repository contains the following tools in the sub directories:

- [rivers](./cmd/rivers): Simple TCP reverse proxy for HA control plane.
- [etcdbackup](./cmd/etcdbackup): Simple etcd backup service.
- [scripts](./scripts): Utilities

License
-------

CKE tools is licensed under MIT license.

[godoc]: https://godoc.org/github.com/cybozu/neco-containers/cke-tools/src
[CKE]: https://github.com/cybozu-go/cke
