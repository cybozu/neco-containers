# bucket-provisioner-light

A minimal [lib-bucket-provisioner](https://github.com/kube-object-storage/lib-bucket-provisioner) implementation that handles `ObjectBucketClaim` resources and creates/deletes buckets on an S3-compatible endpoint.

> [!WARNING]
> lib-bucket-provisioner is deprecated upstream.

## Features

- `Provision`: creates a bucket if it does not exist
- `Delete`: removes all objects in the bucket, then deletes the bucket
- `Grant` / `Revoke`: not implemented
- No IAM user or policy management
