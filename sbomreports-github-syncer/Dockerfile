FROM ghcr.io/cybozu/golang:1.26-noble AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/sbomreports-github-syncer .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/sbomreports-github-syncer /usr/local/bin/sbomreports-github-syncer
USER 65532:65532
ENTRYPOINT ["/usr/local/bin/sbomreports-github-syncer"]
