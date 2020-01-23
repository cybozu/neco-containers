Configuration
=============

The configuration of `neco-admission` is a collection of webhooks configurations.
This collection is indexed by webhooks names.

ArgoCDApplicationValidator
-------------------------

The configuration of `ArgoCDApplicationValitor` is a map with the following keys.

| Name  | Type     | Description                                |
| ----- | -------- | ------------------------------------------ |
| rules | \[\]rule | A list of rules to enforce `spec.project`. |

Each rule represents the restriction on the applications in a certain repository.

| Name       | Type       | Description                                                               |
| ---------- | ---------- | ------------------------------------------------------------------------- |
| repository | string     | A URL of the repository                                                   |
| projects   | \[\]string | A list of `spec.project`s allowed for the applications in the repository. |

```yaml
ArgoCDApplicationValidator:
  rules:
    - repository: https://github.com/cybozu-private/maneki-apps.git
      projects:
        - maneki
```
