# sabakan container

# Stage1: build from source
FROM quay.io/cybozu/golang:1.10-bionic AS build

RUN go get -u github.com/cybozu-go/sabakan/...

RUN apt-get update \
    && apt-get -y install --no-install-recommends grub-ipxe \
    && rm -rf /var/lib/apt/lists/*

# Stage2: setup runtime container
FROM quay.io/cybozu/ubuntu:18.04

COPY --from=build /go/bin /usr/local/sabakan/bin
COPY --from=build /go/src/github.com/cybozu-go/sabakan/LICENSE /usr/local/sabakan/LICENSE
COPY --from=build /boot/ipxe.efi /usr/lib/ipxe/ipxe.efi
COPY --from=build /usr/share/doc/grub-ipxe/copyright /usr/share/doc/grub-ipxe/copyright

ENV PATH=/usr/local/sabakan/bin:"$PATH"

# DHCP and HTTP
EXPOSE 10067 10080

USER 10000:10000

ENTRYPOINT ["/usr/local/sabakan/bin/sabakan", "-url-port", "80"]
