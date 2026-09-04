---
name: update-container
description: >-
  Update one container image in neco-containers to a new upstream version.
  Resolves the image directory, checks the upstream release, follows the image's
  own maintenance.md procedure, bumps TAG/BRANCH, and verifies the image builds
  exactly the way CI does. Use whenever asked to update, bump or upgrade a
  specific container image, or for コンテナ定期更新 / Regular Update /
  Kubernetes Update work in this repository.
---

# Update a container image

Input: **one image name** (e.g. `alertmanager`, `etcd`, `golang-1.26-noble`),
optionally with a target version. Without a target version, look up the latest
upstream release.

Do one image per invocation. For a batch (定期更新), run this once per image.

The authority for *what* to edit is always the image's own section in
`maintenance.md`. This file is the surrounding workflow: how to find that
section, how to decide whether to update at all, how to name the tag, and how to
prove the image still builds.

## Step 1 — Resolve the image to a directory

```bash
grep -rl 'container-image: "<name>"' --include=build-targets.yaml .
```

The image name and the directory name are not always the same
(`golang-all/golang-1.26-noble` publishes `ghcr.io/cybozu/golang`). Directories
nest at most one level deep — `generate_matrix` searches `-mindepth 2 -maxdepth 3`.

Special cases:

- **`ceph/` and `envoy/`** have no `build-targets.yaml`. They are built by
  `.github/actions/build_ceph` / `build_envoy` on every CI run, using
  `target.yaml` in those action directories. `scripts/verify_build.sh` does not
  handle them, and both builds take a very long time (Bazel / `ceph/build.sh`).
  Ask the user before attempting a local build.
- **`golang-all/*`** is bumped automatically by the `Update Go Versions` cron in
  `.github/workflows/update.yaml`. Say so before updating it by hand.

## Step 2 — Read the image's maintenance.md section

Read from `## <image>` to the next `##` in `maintenance.md`. **Follow those
numbered steps literally**; they name the exact variables and files to touch.

Check the badge at the top of the section:

| Badge | Meaning |
| --- | --- |
| `Regular Update` | Quarterly update — in scope. |
| `Kubernetes Update` | Updated with the Kubernetes upgrade — in scope for that campaign. |
| `CSA Update` | Owned by the CSA team (ceph, cephcsi, rook, csi sidecars, local-pv-provisioner …). Ask before continuing. |
| `No Need Update` | PoC only (falco family, trivy-operator). Ask before continuing. |

Per the regular-update procedure, some images are also owned by other teams even
when they carry a Regular Update badge: **Argo CD, golang, Teleport, Accurate,
Cattage** are PDX's; **ECK** is Cloud Platform's; **MOCO** is DBRE's. Mention
this rather than silently updating them.

## Step 3 — Decide whether, and to what, to update

Current version: `<dir>/TAG` plus the Dockerfile `ARG`/`ENV` variables
(`*_VERSION`, `*_HASH`, `*_SHA256`, `*_COMMIT`).

Latest upstream: use the release page linked in the maintenance.md section.

```bash
gh release list -R <owner>/<repo> -L 20
gh api repos/<owner>/<repo>/releases/latest --jq '.tag_name'
```

Projects not on GitHub (bird, squid, chrony …) list their own URL in the
section — fetch that.

Then apply the regular-update policy:

- **Major version bump → do not do it here.** Report it and recommend a separate
  issue. If a minor release also exists on the current major, do that one.
- **Minor/patch with a breaking change → stop.** Report it and recommend a
  separate issue.
- **Minor/patch with no breaking change → proceed.**
- **No upstream change, but a rebuild is wanted** (e.g. refreshing the Ubuntu
  base image) → bump only the last component of `TAG`.

Skim the release notes for things worth reporting — breaking changes, metric
name/label changes, bundled CLI tool versions, patches now merged upstream — but
do not go deep. CI, neco-apps CI and the stage cluster are what actually verify
the change.

## Step 4 — Apply the changes

Follow the maintenance.md steps. Recurring patterns:

- **Version + hash variables in the Dockerfile.** Recompute hashes for real,
  e.g. `curl -sSL <url> | sha256sum`. Never carry over an old hash.
- **Images that derive the version from `TAG`** (`COPY TAG /` plus
  `cut -d . -f1-3 < /TAG`) need no version edit in the Dockerfile — only `TAG`
  and any `*_HASH`. These include `approver-policy`, `trust-packages`,
  `trust-manager`, `spire-*`, `hubble*`, `cilium-certgen`,
  `cilium-operator-generic`, `stakater-reloader`.
- **Base images** (`ghcr.io/cybozu/ubuntu:24.04`, `ghcr.io/cybozu/golang:1.26-noble`).
- **Patch files.** Check whether each `*.patch` / `*.diff` is now included
  upstream; if so delete the file and its `RUN patch` line.
- **Go components in this repo**: update the direct dependencies in `go.mod`,
  then `go mod tidy`. Unless there is a reason not to, update all dependent
  modules. Run `aqua update-checksum --prune` where the section says to.

  **Leave Kubernetes-version-dependent modules alone.** They are pinned to the
  cluster's Kubernetes version and are bumped only by the `Kubernetes Update`
  campaign, which sets them explicitly to the versions in Kubernetes' own
  `go.mod`. Bumping them during a regular update desynchronises them from the
  cluster. Do not touch:

  - `k8s.io/*` — `api`, `apimachinery`, `client-go`, `component-base`,
    `apiextensions-apiserver`, `kubectl`, `cli-runtime`, …
  - `sigs.k8s.io/controller-runtime` and `sigs.k8s.io/controller-tools`
  - the `go` directive / toolchain version in `go.mod`
  - `ENVTEST_K8S_VERSION`, `KUBERNETES_VERSION`, `E2ETEST_K8S_VERSION` and
    kind / kubectl entries in `Makefile` or `aqua.yaml`

  So run `go get` on the remaining direct dependencies one by one (or by name)
  rather than `go get -u ./...`, and after `go mod tidy` diff `go.mod` to confirm
  no `k8s.io/*`, `controller-runtime` or `go` directive line moved. If a
  non-Kubernetes dependency cannot be updated without dragging one of those
  along, stop and report it instead of forcing it.
- **`README.md` image tag**, when the section lists that step.
- **Cross-container follow-ups** named in the section — e.g. argocd also drives
  `ARGOCD_VERSION` in `admission/Makefile` and the dex / redis / haproxy images.

### TAG

Rules live in the repo `README.md` ("Tag naming"). Format is
`<upstream version>.<container image version>`.

- Upstream changed → `<new upstream>.1`. **The trailing component must reset to 1.**
- Rebuild only → increment the trailing component (`.2`, `.3`, …).
- Upstream `X.Y` with no patch part → pad: `X.Y.0.A`.
- Debian package upstream `X.Y.Z-PACKAGE` → `X.Y.Z.PACKAGE.A`.
- Images with no upstream (neco-admission etc.) use plain semver `X.Y.Z`.
- Preserve suffixes: `1.26.8.1_noble`, `0.26.0.2-ubuntu22.04`.

### BRANCH

Only where a `BRANCH` file exists. For upstream `X.Y.Z` the branch is `X.Y`, or
`0` when the major is 0. Leave alone any entry listed in
`NO_TAG_BRANCH_CONSISTENCY` (e.g. `latest`), and do not invent new entries.

Bumping the minor in `TAG` without bumping `BRANCH` makes
`tag_branch_consistency` fail, which fails CI.

## Step 5 — Verify the build

```bash
.claude/skills/update-container/scripts/build_plan.sh <dir>
.claude/skills/update-container/scripts/verify_build.sh <dir>
```

`build_plan.sh` runs `build_decision` with `SKIP_TAG_CHECK=true`, so it needs no
PR and no registry access. It prints one plan JSON per image in the directory and
runs the same `tag_branch_consistency` / `tags_gen` checks CI runs — a non-zero
exit means CI would fail too. **Read the `tags` field and confirm the new
version is there** before building.

`verify_build.sh` then reproduces `.github/actions/build_push`:
`make -C <dir> <make-targets>` → `docker buildx build` → `IMAGE_TAG=<primary tag>
make -C <dir> <make-post-targets>`. Nothing is pushed.

Success = the build finishes and the post-targets pass. Report the built tag.

Watch out for:

- Multi-platform entries (`linux/amd64,linux/arm64`, e.g. golang) need a
  buildx container driver and QEMU: `docker buildx create --use`.
- `cilium` (and the `ceph` / `envoy` special cases) take tens of minutes.
  Confirm with the user before starting.
- CI ignores `.md`-only changes, and **silently skips** the build when the tag
  already exists on ghcr. A forgotten `TAG` bump is not an error — it is a
  no-op. This is why Step 5 starts by checking the tag list.

## Step 6 — Report

- What changed, and the new image tag.
- Release-note highlights: breaking changes, metric changes, CLI tool versions.
- Anything that should become a separate issue (major bump, breaking change).
- Downstream follow-ups implied by the update:
  - `artifacts.go` and `common.mk` in `cybozu-go/neco` — images listed there must
    be reflected promptly or neco CI breaks.
  - manifests in `neco-apps`.
  - `test/crd-schemas` in `neco-apps` for Cilium and Coil (`make update-cilium`,
    `make update-coil`).
  - Neco Release Notes, once the change reaches a real cluster.
  - Network components (Cilium, Coil, Contour/Envoy, Squid) need extra care and
    a gap between stage and prod rollout.
- Suggested branch and commit message, following existing history:
  branch `cybozu/<image>-<version>`, commit `<image>: update to <version>`.

**Do not commit, push, or open a PR unless asked.**
