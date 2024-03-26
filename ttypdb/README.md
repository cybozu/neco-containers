# ttypdb

## ttypdb-controller

### command-line flags

- `-l`, `--selector`: Selector to filter Pods. Same as `-l` of `kubectl get`. default: `""` (empty string. i.e. select all Pods)
- `--interval`: Polling interval in seconds. default: `60`

## ttypdb-sidecar

ttypdb-sidecar counts and exposes the number of controlling terminals.

- Pods must have `shareProcessNamespace: true` specified.
- The name of the ttypdb-sidecar container must be `ttypdb-sidecar`.
- The container must have exactly one port.

### command-line flags

currently, none.
