FROM ghcr.io/cybozu/golang:1.22-jammy AS build

ARG K8S_VERSION="1.28.7"

COPY . /work/src
WORKDIR /work/src
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o daemonset-updater && \
	curl -o ./kubectl https://storage.googleapis.com/kubernetes-release/release/v${K8S_VERSION}/bin/linux/amd64/kubectl && \
	chmod a+x ./kubectl

FROM ghcr.io/cybozu/ubuntu:22.04
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /work/src/daemonset-updater /
COPY --from=build /work/src/kubectl /usr/bin/
ENTRYPOINT ["/daemonset-updater"]
