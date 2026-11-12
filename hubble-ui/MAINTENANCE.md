# hubble-ui

This image is shared by all supported Cilium versions.

1. Check the [releases](https://github.com/cilium/hubble-ui/releases) page for changes.
2. Update the `BRANCH` and `TAG` files accordingly.
3. Update `aqua.yaml` and run `aqua update-checksum -prune`.
4. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/cilium/hubble-ui/blob/vX.Y.Z/Dockerfile` (replace `vX.Y.Z` with the tag from `TAG`)
   - Update `NODE_VERSION` (full patch version, e.g. `22.22.2`) and `NODE_SHA` in `Dockerfile`. Get the SHA256 from `https://nodejs.org/dist/v${NODE_VERSION}/SHASUMS256.txt` for `node-v${NODE_VERSION}-linux-x64.tar.xz`.
   - Update `NGINX_VERSION` in `Dockerfile`.
   - Update `HUBBLE_UI_COMMIT_HASH` in `Makefile`.
   - Update `NGINX_COMMIT_HASH` in `Makefile`.
      - Browse <https://github.com/nginx/docker-nginx-unprivileged/commits/main/> .
      - `NGINX_COMMIT_HASH` should be the one referencing the commit "Update mainline NGINX to <NGINX_VERSION>".
5. Regenerate `pnpm-lock.yaml`. `hubble-ui/.pnpmfile.cjs` pins two lists of packages to the versions in upstream's `package-lock.json` (checked out in `src/hubble-ui`): `PHANTOM`, packages upstream imports but does not declare, and `OVERRIDE`, packages that would otherwise resolve to a version different from upstream's. `make import-lockfile` requires [Takumi Guard](https://shisho.dev/docs/ja/t/guard/quickstart/npm/) configured locally — every new tarball is fetched through the proxy for supply-chain screening, and the target refuses to start otherwise.
   It also resolves with `minimumReleaseAge=4320` (3 days), so a version published less than 3 days ago is skipped in favor of an older one. If upstream requires such a version, wait, or pass `PNPM_MINIMUM_RELEASE_AGE=0` once and note it in the PR.

   ```sh
   cd hubble-ui
   make import-lockfile
   ```

   Re-run it whenever `HUBBLE_UI_COMMIT_HASH` changes. With a stale lockfile, `make verify-lockfile` (also run in CI) and the Docker build fail with `ERR_PNPM_OUTDATED_LOCKFILE`.
   If it fails with `<pkg> is now declared in upstream package.json`, remove `<pkg>` from `PHANTOM`; with `<pkg> not found in upstream package-lock.json`, remove `<pkg>` from whichever list has it.

   If `make build-frontend` fails at `pnpm run build`, edit `hubble-ui/.pnpmfile.cjs` and run `make import-lockfile` again:
   - `Module not found: Can't resolve '<pkg>'`: add `<pkg>` to `PHANTOM`.
   - `instanceof` or type errors from `<pkg>` whose version in `pnpm-lock.yaml` differs from `src/hubble-ui/package-lock.json`: add `<pkg>` to `OVERRIDE`.

   Review the diff:
   - `pnpm-lock.yaml` — check new entries and integrity hashes for anomalies.
   - Packages with install scripts: compare with the previous release and review any new or bumped one before merging. `--ignore-scripts` keeps dependency scripts from running during the build; upstream's own `postinstall` (`patch-package` + `scripts/post-install.sh`) is run on purpose, so review changes to those as well.

     ```sh
     PREV=  # HUBBLE_UI_COMMIT_HASH before this update
     cd src/hubble-ui
     git fetch --depth 1 origin ${PREV}
     diff <(git show ${PREV}:package-lock.json | jq -r '.packages | to_entries[] | select(.key != "" and .value.hasInstallScript) | "\(.key)@\(.value.version)"') \
          <(jq -r '.packages | to_entries[] | select(.key != "" and .value.hasInstallScript) | "\(.key)@\(.value.version)"' package-lock.json)
     ```
6. Check the upstream [Dockerfile](https://github.com/nginx/docker-nginx-unprivileged/blob/main/Dockerfile-debian.template) for unprivileged version of nginx.

   ```sh
   OLD_NGINX_VERSION=
   NEW_NGINX_VERSION=

   diff --side-by-side -W300 \
   <(curl -sL https://raw.githubusercontent.com/nginx/docker-nginx-unprivileged/refs/tags/${OLD_NGINX_VERSION}/Dockerfile-debian.template) \
   <(curl -sL https://raw.githubusercontent.com/nginx/docker-nginx-unprivileged/refs/tags/${NEW_NGINX_VERSION}/Dockerfile-debian.template)

   diff --side-by-side -W300 \
   <(cat hubble-ui/Dockerfile) \
   <(curl -sL https://raw.githubusercontent.com/nginx/docker-nginx-unprivileged/refs/tags/${NEW_NGINX_VERSION}/Dockerfile-debian.template)
   ```

7. Update `NJS_VERSION` and `PKG_RELEASE` in `Dockerfile`.
