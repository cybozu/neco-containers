# redis_exporter container

This directory provides a Dockerfile to build a redis_exporter container
that runs `redis_exporter` in [oliver006/redis_exporter](https://github.com/oliver006/redis_exporter),
a Prometheus exporter for Redis and Redis Sentinel.

## Usage

The exporter scrapes the instance specified by `REDIS_ADDR` and listens on
`:9121` by default.

```bash
docker run -p 9121:9121 -e REDIS_ADDR=redis://localhost:6379 ghcr.io/cybozu/redis_exporter:X.Y
```

Sentinel metrics (`redis_sentinel_*`) are exported automatically when
`REDIS_ADDR` points to a Sentinel port; no dedicated flag is required.

```bash
docker run -p 9355:9355 \
    -e REDIS_ADDR=redis://localhost:26379 \
    -e REDIS_EXPORTER_WEB_LISTEN_ADDRESS=:9355 \
    ghcr.io/cybozu/redis_exporter:X.Y
```

Note that Sentinel usually has no `requirepass`, so `REDIS_PASSWORD` must not be
given for a Sentinel target. Otherwise the AUTH command is rejected and
`redis_up` becomes 0.

## Docker images

Docker images are available on [ghcr.io](https://github.com/cybozu/neco-containers/pkgs/container/redis_exporter)
