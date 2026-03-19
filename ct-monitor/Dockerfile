FROM ghcr.io/cybozu/golang:1.26-noble AS build
ARG VERSION=0.6.5
COPY plugin/ /work/plugin/
RUN CGO_ENABLED=0 go install github.com/Hsn723/ct-monitor@v${VERSION} \
    && mkdir -p /etc/ct-monitor /var/log/ct-monitor /plugins \
    && find /work/plugin -mindepth 1 -maxdepth 1 -type d | while read d; do \
         CGO_ENABLED=0 go build -o /plugins/$(basename "$d") "$d"; \
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

USER 65534:65534
ENTRYPOINT ["/ct-monitor"]
