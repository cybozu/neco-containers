# trust-packages

![Regular Update](../regular_update.svg)

1. Check the [release page](https://github.com/cert-manager/trust-manager/releases).
2. Update `TRUST_MANAGER_HASH` and `UBUNTU_IMAGE` in `Makefile`.
3. Update base image and `TRUST_MANAGER_HASH` in `Dockerfile`.
4. Update `TAG` file.
5. Verify that the new `trust-packages` version matches the `trust-manager` version.
