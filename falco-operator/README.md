Falco Operator container
========================

A single `Dockerfile` builds both operators of [falco-operator](https://github.com/falcosecurity/falco-operator) as two separate images:

- `falco-operator`: the instance operator (`./cmd/instance`), which deploys and manages Falco instances on Kubernetes. Built on `scratch`.
- `falco-artifact-operator`: the artifact operator (`./cmd/artifact`), which manages Falco artifacts such as rules and plugins. Built on `ubuntu`.

The instance operator deploys the artifact operator, so the artifact operator image reference is baked into the instance operator binary at build time via the `ArtifactOperatorImage` ldflag (`ARTIFACT_OPERATOR_IMAGE` build arg).

Docker images
-------------

Docker images are available on [ghcr.io](https://github.com/cybozu/neco-containers/pkgs/container/falco-operator) (`falco-operator`) and [ghcr.io](https://github.com/cybozu/neco-containers/pkgs/container/falco-artifact-operator) (`falco-artifact-operator`).
