# alloy

![Regular Update](../regular_update.svg)

Alloy contains two versions, one for alloy and the other for libsystemd.
The libsystemd version should be the same with the one running on [the stable Flatcar OS](https://www.flatcar.org/releases).

1. Check the release page.
   - <https://github.com/grafana/alloy/releases>
2. Check the upstream Dockerfile and Makefile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/grafana/alloy/blob/vX.Y.Z/.github/workflows/build.yml`
   - `https://github.com/grafana/alloy/blob/vX.Y.Z/Dockerfile`
   - `https://github.com/grafana/alloy/blob/vX.Y.Z/Makefile`
   - `https://github.com/grafana/alloy/blob/vX.Y.Z/tools/build-image/Dockerfile`
3. Update `ALLOY_VERSION` in `Dockerfile`.
4. Update `SYSTEMD_VERSION` in `Dockerfile` if needed.
5. Update `TAG` file.
