# alertmanager

1. Check the release page.
   - <https://github.com/prometheus/alertmanager/releases>
2. Check the upstream Makefile and Dockerfile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/prometheus/alertmanager/blob/vX.Y.Z/Makefile`
   - `https://github.com/prometheus/alertmanager/blob/vX.Y.Z/Dockerfile`
   - `https://github.com/prometheus/alertmanager/blob/vX.Y.Z/ui/Dockerfile` (check Node.js version)
3. Update version variables in `Dockerfile`.
4. Update `BRANCH` and `TAG` files.
