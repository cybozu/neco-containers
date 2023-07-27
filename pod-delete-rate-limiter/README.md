pod-delete-rate-limiter
=======================

pod-delete-rate-limiter is a rate-limiter for Pod deletion.

This program is originally written intended to rate-limit StatefulSet rolling update by rate-limiting Pod deletion with validating webhook.

Options
-------

- `-health-probe-bind-address` The address the probe endpoint binds to. (default `:8081`)
- `-limited-user` The user who is applied rate limit. (default `system:serviceaccount:kube-system:statefulset-controller`)
- `-metrics-bind-address` The address the metric endpoint binds to. (default `:8080`)
- `-min-interval` The minimum interval in seconds for deletion. (default `1.0`)
- and zap logger related options
  - `-zap-devel`
  - `-zap-encoder`
  - `-zap-log-level`
  - `-zap-stacktrace-level`
  - `-zap-time-encoding`
