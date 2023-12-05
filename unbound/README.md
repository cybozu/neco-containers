# Unbound container

[Unbound](https://nlnetlabs.nl/projects/unbound/about/) is a DNS resolver.

## Usage

### Launch Unbound with specific config file

Prepare config file `unbound.conf` at working directory, then execute following command.

    $  docker run --mount type=bind,source="$(pwd)"/unbound.conf,target=/etc/unbound.conf \
        ghcr.io/cybozu/unbound:1.18 -c /etc/unbound.conf
 
## Docker images

Docker images are available on [ghcr.io](https://github.com/cybozu/neco-containers/pkgs/container/unbound)
