# Redis container

[Redis](https://redis.io/) is an in-memory database that persists on disk.

## Usage

### Launch Redis

```console
$ docker run --name=redis quay.io/cybozu/redis:6.2
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

Docker images are available on [Quay.io](https://quay.io/repository/cybozu/redis)
