## Instruction to add a new Collector

neco-exporter is a collection of neco-specific metrics collectors.  
If you want to add a new collector to extend functionality, follow the below instruction.

1. Determine serving scope
    - If the metrics represents cluster property, it should be served when `--scope=cluster`
    - If the metrics represents node property, it should be served when `--scope=node`
2. Determine short name
    - Name your collector with a unique short name, e.g. `bpf`, `ciliumid`, `mock`...
3. Add a collector
    - If yours runs in cluster-scope, add it under `pkg/collector/cluster`
    - If yours runs in node-scope, add it under `pkg/collector/node`
4. Add your collector to bootstrap
    - Open `main.go` and add yours in `factories` variable

### Instruction for CI

1. Open `e2e/testdata/daemonset.yaml` or `e2e/testdata/deployment.yaml` depending on the serving scope
2. Add your collector's short name to `--collectors`
3. Write necessary test
