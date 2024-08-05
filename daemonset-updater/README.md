# daemonset-update

daemonset-update updates the daemonset where the update strategy is OnDelete.

## Usage

```
$ ./daemonset-updater -h
daemnonset-updater updates daemonsets that is on-delete strategy

Usage:
  daemonset-updater [flags]

Flags:
  -l, --app-label string         The label associated with pods that is a part of the daemonset. exp) app=test
  -d, --desired-image string     Desired image
      --drain-timeout duration   Timeout for draining (default 30m0s)
  -h, --help                     help for daemonset-updater
      --ignore strings           List of nodes that is ignored
      --kubectl-path string      Path to kubectl (default "/usr/bin/kubectl")
      --timeout duration         Total timeout to update (default 9h0m0s)
  -v, --version                  version for daemonset-updater
```
