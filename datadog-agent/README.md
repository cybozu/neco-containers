Datadog agent
==================

Build Docker container image for [datadog-agent][https://hub.docker.com/r/datadog/agent].

Usage
-----

### Run gorush:

```console
$ docker run -d --rm --read-only \
    --name datadog-agent \
    quay.io/cybozu/datadog-agent:7.17.0.1
```

[Gorush]: https://github.com/DataDog/datadog-agent

Docker images
-------------

Docker images are available on [Quay.io](https://quay.io/repository/cybozu/datadog-agent)
