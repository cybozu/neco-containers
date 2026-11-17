# How to maintain neco-containers <!-- omit in toc -->

This document describes the procedure for updating each container image.

Besides on-demand container updates, we have two regular renewal operations: `Kubernetes Update` and `Regular Update`.
The target container of these operations have the following badges, so check before the operations.

In case of components whose Go source code are in neco-containers, all dependent Go modules should be updated if there is no special reason. Kubernetes-related modules such as client-go may be newer than the Kubernetes to be updated. For example, it is acceptable that client-go is v0.30 and Kubernetes is v1.29.

1. Kubernetes Update (![Kubernetes Update](./kubernetes_update.svg))
   - Upgrade of Kubernetes. Besides the related components of Kubernetes, update the containers managed by [CKE](https://github.com/cybozu-go/cke/) and some go modules.
2. Regular Update (![Regular Update](./regular_update.svg))
   - Update in every quarter. Keeping up with the upstream version and updating the ubuntu base image.
3. CSA Update  (![CSA Update](./csa_update.svg))
   - Update by CSA team.
4. No Need Update (![No Need Update](./no_need_update.svg))
   - Used as a PoC, so regular updates are not required.
