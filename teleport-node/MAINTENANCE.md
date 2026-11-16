# teleport-node

1. Check the [release page](https://github.com/gravitational/teleport/releases).
2. Run `make -C teleport-node/ check-teleport-update` and check the upstream `Makefile` and `version.mk`.
3. Update tools version in `Dockerfile`.
4. Update `Dockerfile` If there are any changes to the build method.
5. Update `TELEPORT_VERSION` in `Dockerfile`.
6. Update `TAG` files.
