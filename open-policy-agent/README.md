[![Docker Repository on Quay](https://quay.io/repository/cybozu/open-policy-agent/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/open-policy-agent)

Open Policy Agent container
===========================

Build Docker container image for [Open Policy Agent][], policy-based control for cloud native environments.

Usage
-----

### Start `opa`

Run the container

```console
$ docker run -d --read-only --name=opa \
    quay.io/cybozu/open-policy-agent:0.12.0 run --server
```

[Open Policy Agent]: https://www.openpolicyagent.org/
