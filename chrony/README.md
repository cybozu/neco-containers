# chrony-container
Build Docker container image for Chrony NTP server/client

## Features
- Chrony 3.3 based on Ubuntu 18.04.
- Multi-stage build to minimize the container size.

## Usage

### Start `chronyd`

1. Prepare chrony.conf
1. Run the container
    ```console
    $ docker run -d --read-only --name=chrony \
    --mount type=bind,source=/your/chrony.conf,target=/etc/chrony.conf,readonly \
    --mount type=tmpfs,destination=/run/chrony,tmpfs-mode=700 \
    --mount type=tmpfs,destination=/var/lib/chrony,tmpfs-mode=755 \
    --publish=123:123/udp \
    --cap-drop ALL \
    --cap-add SYS_TIME \
    --cap-add NET_BIND_SERVICE \
    quay.io/cybozu/chrony:3.3
    ```

### Use `chronyc`

- Let `chronyd` running at `10.0.0.3`.
- Let settings are written in `chrony.conf` to allow access from the endpoint to run `chronyc`.(only localhost is allowed in default)

1. Run the container
    ```console
     $  docker exec -it chrony chronyc tracking -h 10.0.0.3
    ```

## LICENSE

MIT
