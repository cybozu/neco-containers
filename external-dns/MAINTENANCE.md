# external-dns

1. Check the [release page](https://github.com/kubernetes-sigs/external-dns/releases).
2. Check the upstream `.ko.yaml`. If there are any updates, update our `Dockerfile`.
   - `https://github.com/kubernetes-sigs/external-dns/blob/vX.Y.Z/.ko.yaml`
3. Update base image in `Dockerfile`.
4. Update `EXTERNALDNS_VERSION` and `EXTERNALDNS_HASH` in `Dockerfile`.
5. Update image tag in `README.md`.
6. Update `TAG` file.
