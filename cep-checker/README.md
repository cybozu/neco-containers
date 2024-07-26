# cep-checker

cep-checker checks the consistency between Pod and CiliumEndpoint.

## Usage

```
$ ./cep-checker -h
cep-checker checks missing Pods or CiliumEndpoints

Usage:
  cep-checker [flags]

Flags:
  -h, --help                    help for cep-checker
  -i, --interval duration       Interval to check missing CEPs or Pods (default 30s)
  -m, --metrics-server string   Metrics server address and port (default "0.0.0.0:8080")
  -v, --version                 version for cep-checker
```

## Metrics

```
// Gauge
cep_checker_missing{name="cep name", namespace="namespace", resource="cep"}
cep_checker_missing{name="pod name", namespace="namespace", resource="pod"}
```
