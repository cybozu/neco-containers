FROM ghcr.io/cybozu/golang:1.22-jammy AS build

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

COPY . /work
WORKDIR /work
# Build statically linked binary.
RUN CGO_ENABLED=0 go build ./...

FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/moco-switchover-downtime-monitor /moco-switchover-downtime-monitor
# Note that this container image does not include kubectl and kubectl-moco.
# You have to add them and PATH environment variable in some way.

USER 10000:10000
ENTRYPOINT [ "/moco-switchover-downtime-monitor" ]
