# envoy

> [!Note]
> Envoy is managed by [Contour](../contour/MAINTENANCE.md) so update to the supported version. See the below.
>
> - [Contour Compatibility Matrix](https://projectcontour.io/resources/compatibility-matrix/)

1. Check the [release page](https://github.com/envoyproxy/envoy/releases).
2. Update `envoy_commit_sha` in [`.github/workflows/main.yaml`](/.github/workflows/main.yaml).
3. Update image tag in `README.md`.
4. Upgrade direct dependencies listed in `go.mod`. Use `go get` or your editor's function.
5. Update `BRANCH` and `TAG` files.
