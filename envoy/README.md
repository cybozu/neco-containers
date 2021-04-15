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
    quay.io/cybozu/envoy:1.17
    ```

Livenessprobe
-----

Envoy has its own probe, but it does not guarantee that Envoy is working correctly.
Therefore, we developed a custom probe for confirming Envoy is listening on HTTP/HTTPS endpoints.

As Envoy does not start listening on HTTP/HTTPS endpoints until the corresponding proxy settings are created, the custom probe returns success at the start until it really succeeds to connect.


[Envoy]: https://github.com/envoyproxy/envoy

Docker images
-------------

Docker images are available on [Quay.io](https://quay.io/repository/cybozu/envoy)
