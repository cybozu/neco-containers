[![Docker Repository on Quay](https://quay.io/repository/cybozu/envoy/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/envoy)

Envoy container
====================

Build Docker container image for [Envoy][], cloud-native high-performance edge/middle/service proxy.

Usage
-----

### Start `envoy`

1. Prepare envoy.yaml
2. Run the container
    ```console
    $ docker run -d -p 10000:10000 --read-only --name=envoy \
    --mount type=bind,source=/your/envoy.yaml,target=/etc/envoy/envoy.yaml,readonly \
    --mount type=tmpfs,target=/tmp \ 
    quay.io/cybozu/envoy:1.9
    ```

[Envoy]: https://github.com/envoyproxy/envoy
