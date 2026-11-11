# opentelemetry-collector

![Regular Update](../regular_update.svg)

opentelemetry-collector container consists of three repositories: opentelemetry-collector, opentelemetry-collector-contrib and opentelemetry-collector-releases

1. Check the release pages [main](https://github.com/open-telemetry/opentelemetry-collector/releases) [contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases) [release](https://github.com/open-telemetry/opentelemetry-collector-releases/releases).
2. Check the upstream Dockerfile and builder manifest. If there are any updates, update our `Dockerfile`.
   - `https://github.com/open-telemetry/opentelemetry-collector-releases/blob/vX.Y.Z/distributions/otelcol/Dockerfile`
   - `https://github.com/open-telemetry/opentelemetry-collector-releases/blob/vX.Y.Z/distributions/otelcol/manifest.yaml`
3. Update `OTELCOL_VERSION` and/or `OTELCOL_CONTRIB_VERSION` in `Dockerfile`.
4. Update `BRANCH` and `TAG` files.
