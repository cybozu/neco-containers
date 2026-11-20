# falco-operator

Used as a PoC, so regular updates are not required.

A single `Dockerfile` builds two images from the `falcosecurity/falco-operator` repository:

- `falco-operator`: the instance operator (`./cmd/instance`), built on `scratch`.
- `artifact-operator`: the artifact operator (`./cmd/artifact`), built on `scratch`.

The instance operator deploys the artifact operator as `<registry>/<namespace>/artifact-operator:<tag>`. The registry, namespace and tag are baked into the instance operator binary via the `image.Registry`, `image.Namespace` and `version.ArtifactOperatorTag` ldflags.

1. Check the [release page](https://github.com/falcosecurity/falco-operator/releases).
2. Check the upstream Dockerfile and build files. If there are any updates, update our `Dockerfile`.
   - `https://github.com/falcosecurity/falco-operator/blob/vX.Y.Z/build/Dockerfile`
   - `https://github.com/falcosecurity/falco-operator/blob/vX.Y.Z/Makefile`
3. Update `FALCO_OPERATOR_COMMIT` in `Dockerfile`.
4. Update `ARTIFACT_OPERATOR_TAG` in `Dockerfile` to match the new `TAG`.
5. Update `TAG` file.
