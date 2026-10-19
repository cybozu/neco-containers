# Redis container

[Redis](https://redis.io/) is an in-memory database that persists on disk.

## Usage

### Launch Redis

```bash
docker run --name=redis ghcr.io/cybozu/redis:X.Y
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

## License

For the Redis component in this container image, we choose the AGPLv3 licensing option offered for Redis 8 and later.
A copy of the GNU Affero General Public License v3.0 is included in this repository and in the container image at /usr/local/redis/COPYING.
Corresponding Source for the Redis component is available at:

- Upstream Redis source releases: <https://github.com/redis/redis/releases>
- Build recipe (Dockerfile) and any patches applied by Cybozu: <https://github.com/cybozu/neco-containers/tree/main/redis>

The exact Redis version and the checksum of the source tarball used to build this image are specified in the Dockerfile in this repository. See the `REDIS_VERSION`, `REDIS_DOWNLOAD_URL`, and `REDIS_DOWNLOAD_SHA` build arguments.
