# contour

> [!Note]
> Contour uses Envoy as a "data plane." Keep version correspondence between the contour and [envoy](../envoy/MAINTENANCE.md) images. Check the compatibility matrix below.
>
> - [Contour Compatibility Matrix](https://projectcontour.io/resources/compatibility-matrix/)

1. Check the [release page](https://github.com/projectcontour/contour/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/projectcontour/contour/blob/vX.Y.Z/Dockerfile`
3. Update `CONTOUR_VERSION` and `CONTOUR_SHA` in `Dockerfile`.
4. Update image tag in `README.md`.
5. Update `BRANCH` and `TAG` files.
6. Update `CONTOUR_VERSION` in `admission/Makefile`.
7. Follow maintenance instructions for [neco-admission](../maintenance.md#admission-neco-admission).

> [!Note]
> Contour's HTTPProxy objects are validated by [neco-admission](../maintenance.md#admission-neco-admission).
> You need to update [neco-admission](../maintenance.md#admission-neco-admission) at least for its `Makefile`.
