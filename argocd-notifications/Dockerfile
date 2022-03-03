# Argo CD Notifications container

FROM quay.io/cybozu/golang:1.17-focal as builder

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

ENV REPO=argoproj-labs/argocd-notifications
ENV VERSION=1.2.1

RUN apt-get update && apt-get install ca-certificates

WORKDIR /work
RUN curl -sSLf "https://github.com/${REPO}/archive/v${VERSION}.tar.gz" | \
    tar --strip-components=1 -zxf -
RUN go mod download
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /app/argocd-notifications ./cmd
RUN ln -s /app/argocd-notifications /app/argocd-notifications-backend

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/argocd-notifications /app/argocd-notifications
COPY --from=builder /app/argocd-notifications-backend /app/argocd-notifications-backend
COPY --from=builder /work/LICENSE /LICENSE

USER 10000:10000
