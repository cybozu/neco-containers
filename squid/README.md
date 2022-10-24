# Squid container

[Squid](http://www.squid-cache.org/) is a web proxy cache service.

## Usage

### Run with the default configuration

    $ docker run -d --read-only quay.io/cybozu/squid:5
### Launch Squid with specific config file

Prepare `squid.conf`, then execute following command.

    $  docker run -d --read-only \
        -v /path/to/your/squid.conf:/etc/squid/squid.conf:ro \
        quay.io/cybozu/squid:5

Your `squid.conf` must have the following configurations:

    pid_filename   none
    logfile_rotate 0
    access_log     stdio:/dev/stdout
    cache_log      stdio:/dev/stderr

## Docker images

Docker images are available on [Quay.io](https://quay.io/repository/cybozu/squid)
