# sbomreports-github-syncer

`sbomreports-github-syncer` is a CLI tool that reads Trivy Operator `SbomReport` custom resources from a Kubernetes cluster and commits them to a GitHub repository as JSON files.

It can be run locally or as a Kubernetes CronJob.

## Features

- Lists `SbomReport` resources from all namespaces or a specific namespace.
- Supports Kubernetes label selectors.
- Renders each report's `.report.components` field as a deterministic JSON file.
- Writes an `index.json` file with metadata for all rendered reports.
- Commits all generated files to GitHub in a single commit.
- Optionally deletes stale `.json` files under the configured path prefix.
- Supports GitHub Enterprise via a custom API URL.
- Supports dry-run mode.

## Requirements

- Go 1.26 or later.
- Access to a Kubernetes cluster with Trivy Operator `SbomReport` resources.
- A GitHub token with `Contents: Read and write` permission for the target repository.
- In-cluster RBAC permissions to list `sbomreports.aquasecurity.github.io`.

## Usage

```bash
export KUBECONFIG=$HOME/.kube/config
export GITHUB_TOKEN=github_pat_xxx
export GITHUB_OWNER=your-org
export GITHUB_REPO=sbom-archive
export GITHUB_BRANCH=main
export GITHUB_PATH_PREFIX=clusters/dev/sbomreports
export CLUSTER_NAME=dev

./sbomreports-github-syncer sync --dry-run
./sbomreports-github-syncer sync
```

## Examples

List reports from all namespaces:

```bash
./sbomreports-github-syncer sync
```

List reports from one namespace:

```bash
./sbomreports-github-syncer sync --namespace default
```

Filter reports by label selector:

```bash
./sbomreports-github-syncer sync \
  --selector 'trivy-operator.resource.kind=Deployment'
```

Write files under a custom repository path:

```bash
./sbomreports-github-syncer sync \
  --path-prefix clusters/prod/sbomreports
```

Run without committing to GitHub:

```bash
./sbomreports-github-syncer sync --dry-run
```

## Output layout

Given:

```bash
--path-prefix clusters/dev/sbomreports
```

The repository output will look like this:

```text
clusters/dev/sbomreports/
  index.json
  default/
    example-report.json
  kube-system/
    another-report.json
```

Each report file is written as:

```text
<path-prefix>/<cluster-name>/<namespace>/<report-name>.json
```

## Configuration

All options can be configured with flags or environment variables.

| Environment variable | Flag | Required | Default | Description |
| --- | --- | --- | --- | --- |
| `KUBECONFIG` | `--kubeconfig` | no | empty | Path to kubeconfig. Empty means in-cluster config first, then `$HOME/.kube/config`. |
| `NAMESPACE` | `--namespace` | no | empty | Namespace to list `SbomReport` resources from. Empty means all namespaces. |
| `LABEL_SELECTOR` | `--selector` | no | empty | Kubernetes label selector for `SbomReport` resources. |
| `GITHUB_OWNER` | `--github-owner` | yes, unless `--dry-run` | empty | GitHub repository owner. |
| `GITHUB_REPO` | `--github-repo` | yes, unless `--dry-run` | empty | GitHub repository name. |
| `GITHUB_BRANCH` | `--github-branch` | no | `main` | GitHub branch to update. |
| `GITHUB_API_URL` | `--github-api-url` | no | `https://api.github.com` | GitHub API URL. |
| `CLUSTER_NAME` | `--cluster-name` | no | empty | Cluster name included in `index.json`. |
| `GITHUB_PATH_PREFIX` | `--path-prefix` | no | `sbomreports` | Directory prefix in the target repository. |
| `COMMIT_MESSAGE` | `--commit-message` | no | generated | Commit message. Empty means a timestamped message is generated. |
| `DELETE_MISSING` | `--delete-missing` | no | `false` | Delete stale `.json` files under the path prefix. |
| `FAIL_IF_EMPTY` | `--fail-if-empty` | no | `false` | Return an error if no reports are found. |
| none | `--dry-run` | no | `false` | Render reports and print the file list without calling GitHub. |

## Deleting stale files

By default, stale deletion is disabled.

When `DELETE_MISSING=true` or `--delete-missing` is set, existing `.json` files under `GITHUB_PATH_PREFIX` are deleted if they are not generated from the current Kubernetes result set.

Use this carefully. For first-time runs, start with:

```bash
export DELETE_MISSING=false
./sbomreports-github-syncer sync --dry-run
```

After confirming the generated file list, enable stale deletion if needed.

## Kubernetes deployment

A sample CronJob manifest is available at:

```text
deploy/cronjob.yaml
```

Before applying it, update at least:

- GitHub owner, repository, and branch
- path prefix
- cluster name
- container image
- schedule
- stale deletion setting

Create the namespace and GitHub token secret:

```bash
kubectl create namespace trivy-system

kubectl create secret generic sbomreports-github-syncer \
  --namespace trivy-system \
  --from-literal=github-token="$GITHUB_TOKEN"
```

Apply the CronJob manifest:

```bash
kubectl apply -f deploy/cronjob.yaml
```

Trigger a manual run:

```bash
kubectl create job \
  --from=cronjob/sbomreports-github-syncer \
  --namespace trivy-system \
  sbomreports-github-syncer-manual-$(date +%s)
```

Check logs:

```bash
kubectl logs \
  --namespace trivy-system \
  job/<job-name>
```

## RBAC

The tool needs permission to list `SbomReport` resources:

```yaml
apiGroups: ["aquasecurity.github.io"]
resources: ["sbomreports"]
verbs: ["get", "list"]
```

If namespace labels are included in the index, the service account also needs permission to get namespaces:

```yaml
apiGroups: [""]
resources: ["namespaces"]
verbs: ["get"]
```

## Troubleshooting

### `GITHUB_TOKEN is required`

Set `GITHUB_TOKEN`. This is not required for `--dry-run`.

### `GITHUB_OWNER/GITHUB_REPO or --github-owner/--github-repo are required`

Set the target repository owner and repository name.

### `no SbomReports found`

Check that Trivy Operator is installed and that `SbomReport` resources exist:

```bash
kubectl get sbomreports -A
```

Also check the namespace, label selector, and RBAC permissions.

### GitHub returns a truncated tree

When stale deletion is enabled, the tool reads the repository tree to find existing `.json` files under the path prefix.
If GitHub returns a truncated recursive tree, the tool refuses to delete files to avoid unsafe cleanup.

Use a smaller dedicated repository or a smaller path layout if this happens.
