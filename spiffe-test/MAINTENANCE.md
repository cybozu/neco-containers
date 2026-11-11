# spiffe-test

1. Update the base image in `Dockerfile`. Pin it by the full tag and digest (e.g. `ghcr.io/cybozu/golang:X.Y.Z.N_noble@sha256:...`).
   - When the Go minor version changes, also update the `go` directive in `go.mod`.
2. Update Go dependencies in the `spiffe-test` directory.
3. Update the tool versions in `aqua.yaml` if necessary, and run `aqua update-checksum -prune`.
   - golangci-lint must support the Go version used in `go.mod`.
4. Update the plugin versions in `buf.gen.yaml` to match `google.golang.org/protobuf` and `google.golang.org/grpc`, then run `make proto`.
5. Run `make check` to confirm that lint, generated code, and tests pass.
6. Update `TAG` and `BRANCH` files.
