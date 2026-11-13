# admission (neco-admission)

## Kubernetes Update

1. Update the following version variables in `Makefile`.
   - `ENVTEST_K8S_VERSION`
2. Update the following packages in `aqua.yaml`.
   - `kubernetes-sigs/controller-tools/controller-gen`
   - `kubernetes-sigs/controller-runtime/setup-envtest`
3. Upgrade direct dependencies listed in `go.mod`. Use `go get` or your editor's function.
4. Generate code and manifests.

   ```bash
   cd $GOPATH/src/github.com/cybozu/neco-containers/admission
   make generate manifests
   # Commit, if there are any updated files.
   ```

5. Confirm build and test are green.

   ```bash
   make build test
   ```

6. Update `TAG` file.

## Regular Update

Do the following as part of the update of each CRD-providing product:

1. Update the matching commit hash variables and comments from the following in `Makefile`.
   - `CONTOUR_COMMIT_HASH`
   - `ARGOCD_COMMIT_HASH`
   - `GRAFANA_OPERATOR_COMMIT_HASH`
2. Modify the code to match the new CRDs if CRDs are changed.
   - The code which depended on the CRDs are in the [hook](https://github.com/cybozu/neco-containers/tree/main/admission/hooks) directory.
   - And let's use `Unstructured` instead of use golang library. Take a look at [this PR](https://github.com/cybozu/neco-containers/pull/339/files).
3. Generate code and manifests.

   ```bash
   cd $GOPATH/src/github.com/cybozu/neco-containers/admission
   make clean
   make generate manifests
   # Commit, if there are any updated files.
   ```

4. Confirm build and test are green.

   ```bash
   make build test
   ```

5. Update `TAG` file.

> [!Note]
> Update `Makefile` even if you find no changes in CRDs.
> Update `TAG` if you have changed mutation/validation code.
