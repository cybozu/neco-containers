Rook container
==============

This container uses a [rook][] with the patch to enable raw mode OSD on LV-backed PVC (`6184.patch`).

This patch contains changes in this [PR][]. Please remove this patch after it will be merged for upstream releases. And should maintain the patch when updating the Rook's docker image.

[rook]: https://github.com/rook/rook
[PR]: https://github.com/rook/rook/pull/6184

Docker images
-------------

Docker images are available on [Quay.io](https://quay.io/repository/cybozu/rook)
