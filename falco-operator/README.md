Falco Operator container
========================

A single `Dockerfile` builds both operators of [falco-operator](https://github.com/falcosecurity/falco-operator) as two separate images:

- `falco-operator`: the instance operator (`./cmd/instance`), which deploys and manages Falco instances on Kubernetes. Built on `scratch`.
- `artifact-operator`: the artifact operator (`./cmd/artifact`), which manages Falco artifacts such as rules and plugins. Built on `scratch`.

The instance operator deploys the artifact operator as `<registry>/<namespace>/artifact-operator:<tag>`, where the image name `artifact-operator` is fixed by upstream. The registry, namespace and tag are baked into the instance operator binary at build time via the `image.Registry`, `image.Namespace` and `version.ArtifactOperatorTag` ldflags (`IMAGE_REGISTRY`, `IMAGE_NAMESPACE` and `ARTIFACT_OPERATOR_TAG` build args), so that it refers to `ghcr.io/cybozu/artifact-operator`.

Docker images
-------------

Docker images are available on [ghcr.io](https://github.com/cybozu/neco-containers/pkgs/container/falco-operator) (`falco-operator`) and [ghcr.io](https://github.com/cybozu/neco-containers/pkgs/container/artifact-operator) (`artifact-operator`).
