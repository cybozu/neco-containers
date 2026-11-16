# sealed-secrets

1. Check the [release page](https://github.com/bitnami-labs/sealed-secrets/releases).
2. Check the upstream Dockerfile and compare with ours especially on the runtime stage. If there are any updates, update our `Dockerfile`.
    - `https://github.com/bitnami-labs/sealed-secrets/blob/vX.Y.Z/docker/controller.Dockerfile`
    - `https://github.com/bitnami-labs/sealed-secrets/blob/vX.Y.Z/docker/kubeseal.Dockerfile`
3. Update `SEALED_SECRETS_VERSION` and `SEALED_SECRETS_COMMIT` in `Dockerfile`.
4. Update `TAG` files.
