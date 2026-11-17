# spegel

1. Check the [release page](https://github.com/spegel-org/spegel/releases).
2. Update `SPEGEL_COMMIT` in `Dockerfile`.
3. Update the base image digest in `Dockerfile`.
   - You can get the digest with `docker buildx imagetools inspect gcr.io/distroless/static:latest`.
4. Update `TAG` file.
