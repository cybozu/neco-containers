# falco

![No Need Update](../no_need_update.svg)

Used as a PoC, so regular updates are not required.

1. Check the [release page](https://github.com/falcosecurity/falco/releases).
2. Check the upstream Dockerfile and build files. If there are any updates, update our `Dockerfile`.
   - `https://github.com/falcosecurity/falco/blob/X.Y.Z/docker/falco/Dockerfile`
3. Update `FALCO_VERSION` and `FALCO_COMMIT` in `Dockerfile`. Update `BPFTOOL_COMMIT` if needed.
4. Update `TAG` file.
