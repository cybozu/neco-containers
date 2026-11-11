# serf

![Regular Update](../regular_update.svg)

1. Check the [release page](https://github.com/hashicorp/serf/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/hashicorp/serf/blob/vX.Y.Z/scripts/serf-builder/Dockerfile`
3. Update `SERF_COMMIT` in `Dockerfile` with the commit hash of the new release tag.
4. Update image tag in `README.md`.
5. Update `BRANCH` and `TAG` files.
