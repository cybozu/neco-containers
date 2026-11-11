# grafana

![Regular Update](../regular_update.svg)

1. Check the [release page](https://github.com/grafana/grafana/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/grafana/grafana/blob/vX.Y.Z/Dockerfile`
   - Check `JS_IMAGE` in the Dockerfile
3. Update `GRAFANA_VERSION` in `Dockerfile`.
4. Update installation of Node.js in `Dockerfile` according to `JS_IMAGE` if necessary.
5. Update `TAG`.
