[![Docker Repository on Quay](https://quay.io/repository/cybozu/confd/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/confd)

# confd container

[confd][] -- Manage local application configuration files using templates and data from etcd or consul

## Usage

To launch confd with specific config file.

    $ docker run --rm -it -v /path/to/etc/confd:/etc/confd -v /path/to/kvs.yml:/tmp/kvs.yml \
        quay.io/cybozu/confd:0.16 -onetime -backend file -file /tmp/kvs.yml

[confd]: https://github.com/kelseyhightower/confd
