# Redis container

[Redis](https://redis.io/) is an in-memory database that persists on disk.

## Usage

### Launch Redis

```bash
docker run --name=redis ghcr.io/cybozu/redis:7.0
```

### Run Redis CLI

```console
$ docker exec -it redis redis-cli
127.0.0.1:6379> SET foo bar
OK
127.0.0.1:6379> keys *
1) "foo"
```

## Docker images

Docker images are available on [ghcr.io](https://github.com/cybozu/neco-containers/pkgs/container/redis)
