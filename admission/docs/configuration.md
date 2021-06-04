Configuration
=============

The configuration of `neco-admission` is a collection of webhooks configurations.
This collection is indexed by webhooks names.

ArgoCDApplicationValidator
-------------------------

The configuration of `ArgoCDApplicationValidator` is a map with the following keys.

| Name  | Type     | Description                                |
| ----- | -------- | ------------------------------------------ |
| rules | \[\]rule | A list of rules to enforce `spec.project`. |

Each rule represents the restriction on the applications in a certain repository.  
If neco-admission has no rule for a given App's repoURL, neco-admission denies the API request.

| Name             | Type       | Description                                                                             |
| ---------------- | ---------- | --------------------------------------------------------------------------------------- |
| repository       | string     | A URL of the repository to be matched with `applications.spec.source.repoURL`.          |
| repositoryPrefix | string     | A URL prefix of the repositories to be matched with `applications.spec.source.repoURL`. |
| projects         | \[\]string | A list of `applications.spec.project`s allowed for the applications in the repository.  |

### `.git` suffix in `repository`

In GitHub, `.git` suffix is set at repository URL automatically. However, this suffix is optional. In fact, you can access the repository without the suffix.
In view of this, neco-admission compares the remote URL ignoring `.git` suffix.

### Example

```yaml
ArgoCDApplicationValidator:
  rules:
    - repository: https://github.com/cybozu-private/maneki-apps.git
      projects:
        - maneki
```
