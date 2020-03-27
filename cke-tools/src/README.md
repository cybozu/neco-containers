[![GoDoc](https://godoc.org/github.com/cybozu/neco-containers/cke-tools/src?status.svg)][godoc]
[![Go Report Card](https://goreportcard.com/badge/github.com/cybozu/neco-containers/cke-tools/src)](https://goreportcard.com/report/github.com/cybozu-go/cke-tools)

cke-tools
=========

CKE tools is a suite of the various tools used by [CKE][].
This repository contains the following tools in the sub directories:

- [rivers](./cmd/rivers): Simple TCP reverse proxy for HA control plane.
- [etcdbackup](./cmd/etcdbackup): Simple etcd backup service.
- [updateblock117](./cmd/updateblock117): Fix block device paths to upgrade Kubelet to 1.17 without draining Node.
- [scripts](./scripts): Utilities

License
-------

CKE tools is licensed under MIT license.

[godoc]: https://godoc.org/github.com/cybozu/neco-containers/cke-tools/src
[CKE]: https://github.com/cybozu-go/cke

`updateblock117`
----------------

`updateblock117` is the one-shot program used by CKE when upgrading Kubelet from 1.16 to 1.17.

`updateblock117` has two subcommands.

### `updateblock117 need-update <block-pv-name>`

Check that we should modify the path of the target device file or not.

`block-pv-name` is the name of the PersistentVolume object.

- Returns stdout `{result: "yes"}` if the device needs to tweak.
- Returns stdout `{result: "no"}` if the device needs to tweak.

### `updateblock117 operate <block-pv-name>`

Modify device file path and its symbolic links for Kubelet 1.17.

`block-pv-name` is the name of the PersistentVolume object.

- Returns stdout `{result: "completed"}` if the process has been finished successfully.
