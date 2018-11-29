[![Docker Repository on Quay](https://quay.io/repository/cybozu/squid/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/squid)

# Squid container

[Squid](http://www.squid-cache.org/) is a web proxy cache service.

## Usage

### Launch Squid with specific config file

These log files are symlinked to `/dev/stdout`

```
# ls -l /var/log/squid/
lrwxrwxrwx 1 root root 11 Nov 29 03:22 access.log -> /dev/stdout
lrwxrwxrwx 1 root root 11 Nov 29 03:22 cache.log -> /dev/stdout
lrwxrwxrwx 1 root root 11 Nov 29 03:22 store.log -> /dev/stdout
```

To output log messages, set `access_log` in your configuration file as follows.

```
access_log stdio:/var/log/squid/access.log
cache_log /var/log/squid/cache.log
cache_store_log stdio:/var/log/squid/store.log
```

Prepare configuration file `squid.conf` at working directory, then execute following command.

    $  docker run --mount type=bind,source="$(pwd)"/squid.conf,target=/etc/squid/squid.conf \
        quay.io/cybozu/squid:3.5.27-1-1
