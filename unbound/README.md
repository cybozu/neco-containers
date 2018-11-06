[![Docker Repository on Quay](https://quay.io/repository/cybozu/unbound/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/unbound)

# Unbound container

[Unbound](https://nlnetlabs.nl/projects/unbound/about/) is a DNS resolver.

## Usage

### Launch Unbound with specific config file

Prepare config file `unbound.conf` at working directory, then execute following command.

    $  docker run --mount type=bind,source="$(pwd)"/unbound.conf,target=/etc/unbound.conf \
        quay.io/cybozu/unbound:1.8.1 -c /etc/unbound.conf
