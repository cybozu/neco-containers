[![Docker Repository on Quay](https://quay.io/repository/cybozu/squid/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/squid)

# Squid container

[Squid](http://www.squid-cache.org/) is a web proxy cache service.

## Usage

### Launch Squid with specific config file

Prepare configuration file `squid.conf` at working directory, then execute following command.

    $  mkdir -m 777 /tmp/squid
    $  docker run -it \
        --mount type=bind,source=/tmp/squid,target=/var/log/squid \
        --mount type=bind,source="$(pwd)"/squid.conf,target=/etc/squid/squid.conf \
        quay.io/cybozu/squid:3.5.27.1.9
