# bmc-reverse-proxy

1. Upgrade direct dependencies listed in `go.mod`.
2. Update the Go version.
   - Update the `go` directive in `go.mod`.
   - Update the build image in `Dockerfile` to the latest [`ghcr.io/cybozu/golang`](https://github.com/cybozu/neco-containers/pkgs/container/golang) image for the Go version, and pin it by tag and digest.
   - You can also get the digest with `docker buildx imagetools inspect <image>:<tag>`.
3. Confirm tests are green.

   ```bash
   make setup
   make test
   ```

4. Update the image tags in `bmc-reverse-proxy.yaml`.
   - Both `bmc-reverse-proxy` and `update-resolve-map` use the same image.
5. Update `TAG` file.
