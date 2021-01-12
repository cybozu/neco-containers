Rook container
==============

This container uses a [rook][].

`208.patch` resolves the issue that OB can not be created from OBC using ArgoCD.
After the [PR][] is merged and then Rook uses the fixed library, remove the patch.

[rook]: https://github.com/rook/rook
[PR]: https://github.com/kube-object-storage/lib-bucket-provisioner/pull/208

Docker images
-------------

Docker images are available on [Quay.io](https://quay.io/repository/cybozu/rook)
