# pause

1. Check the changelog.
   - `https://github.com/kubernetes/kubernetes/blob/vX.Y.Z/build/pause/CHANGELOG.md`
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/kubernetes/kubernetes/blob/vX.Y.Z/build/pause/Dockerfile`
3. Update `K8S_COMMIT` and `PAUSE_VERSION` in `Dockerfile`.
4. Update `BRANCH` and `TAG` files.
