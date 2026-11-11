# cadvisor

1. Check the [release page](https://github.com/google/cadvisor/releases).
2. Check the upstream build files. If there are any updates, update our `Dockerfile`.
   - `https://github.com/google/cadvisor/blob/vX.Y.Z/Makefile`
   - `https://github.com/google/cadvisor/blob/vX.Y.Z/build/release.sh`
   - `https://github.com/google/cadvisor/blob/vX.Y.Z/build/build.sh`
   - `https://github.com/google/cadvisor/blob/vX.Y.Z/deploy/Dockerfile`
3. Update `CADVISOR_COMMIT` in `Dockerfile` with the commit hash of the new release tag.
4. Update `TAG` file.
