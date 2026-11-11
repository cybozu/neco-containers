# tempo

1. Check the [release page](https://github.com/grafana/tempo/releases).
2. Check the upstream `Makefile` and `cmd/tempo//Dockerfile`. If they have been updated significantly, update our `Dockerfile`.
   - `https://github.com/grafana/tempo/blob/vX.Y.Z/Makefile`
   - `https://github.com/grafana/tempo/blob/vX.Y.Z/cmd/tempo/Dockerfile`
3. Update `TEMPO_VERSION` in `Dockerfile`.
4. Update `TAG` file.
