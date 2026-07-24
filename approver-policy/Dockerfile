# approver-policy container

# Stage1: build from source
FROM ghcr.io/cybozu/golang:1.26.5.2_noble@sha256:419e350acc85bb2603c45e15e12c3238a1ec765cb58f5eb14ef7adf8bea9748f AS build
ENV APPROVER_POLICY_HASH=6e921ad621d9c8b40638b3556c30f983b31f322ad3bc7c186124a1d6e61373f5

COPY TAG /
WORKDIR /workspace
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN VERSION=$(cut -d \. -f1-3 < /TAG ) \
    && curl -fsSL https://github.com/cert-manager/approver-policy/archive/v${VERSION}.tar.gz -o /tmp/approver-policy.tar.gz \
    && echo "${APPROVER_POLICY_HASH} /tmp/approver-policy.tar.gz" | sha256sum -c - \
    && tar zxf /tmp/approver-policy.tar.gz --strip-components=1 -C /workspace/

RUN CGO_ENABLED=0 GO_EXPERIMENT= go build -ldflags '-w -s' -o bin/approver-policy ./cmd

# Stage2: setup runtime container
FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"
LABEL description="approver-policy is a cert-manager approver for CertificateRequestPolicy-based governance of certificate requests"

WORKDIR /
COPY --from=build /workspace/bin/approver-policy /usr/bin/approver-policy
COPY --from=build /workspace/LICENSE /LICENSE

USER 10000:10000
EXPOSE 6060 6443 9402
ENTRYPOINT ["/usr/bin/approver-policy"]
