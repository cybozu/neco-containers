Falco Operator container
========================

This image builds both operators of [falco-operator](https://github.com/falcosecurity/falco-operator) from a single source:

- `/usr/bin/falco-operator`: the instance operator (`./cmd/instance`), which deploys and manages Falco instances on Kubernetes.
- `/usr/bin/falco-artifact-operator`: the artifact operator (`./cmd/artifact`), which manages Falco artifacts such as rules and plugins.

The image does not set an `ENTRYPOINT`; specify the command to run the desired operator.

Docker images
-------------

Docker images are available on [ghcr.io](https://github.com/cybozu/neco-containers/pkgs/container/falco-operator)
