FROM ghcr.io/cybozu/golang:1.27.3.2_resolute@sha256:2e42484cb145c45c51065e58a62420b3afa2164ed303062ceb2f7b82e978df20 AS build
ARG VERSION=0.6.10
COPY plugin/ /work/plugin/
RUN CGO_ENABLED=0 go install github.com/Hsn723/ct-monitor@v${VERSION} \
    && mkdir -p /etc/ct-monitor /var/log/ct-monitor /plugins /tmp \
    && find /work/plugin -mindepth 1 -maxdepth 1 -type d | while read d; do \
         (cd "$d" && CGO_ENABLED=0 go build -o /plugins/$(basename "$d") .); \
       done

FROM scratch
LABEL org.opencontainers.image.source="https://github.com/cybozu/neco-containers"

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/group /etc/group
COPY --from=build --chown=nobody:nogroup /etc/ct-monitor /etc/ct-monitor
COPY --from=build --chown=nobody:nogroup /var/log/ct-monitor /var/log/ct-monitor
COPY --from=build /go/bin/ct-monitor /
COPY --from=build /plugins /plugins
COPY --from=build --chown=nobody:nogroup /tmp /tmp

USER 65534:65534
ENTRYPOINT ["/ct-monitor"]
