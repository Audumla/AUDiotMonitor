# Stage 1: build — always runs on the host platform (amd64) to avoid QEMU
# overhead. Go's native cross-compilation handles the target arch instead.
FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS builder
ARG TARGETARCH
ARG VERSION=dev
ARG BUILDTIME=unknown
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build \
    -trimpath \
    -ldflags="-s -w -X hwexp/internal/version.Version=${VERSION} -X hwexp/internal/version.BuildTime=${BUILDTIME}" \
    -o /hwexp ./cmd/hwexp/main.go

# Stage 2: runtime
FROM alpine:3
RUN apk add --no-cache tzdata pciutils dmidecode smartmontools
COPY --from=builder /hwexp /usr/bin/hwexp
# Bake in default config — works out of the box with no mounted volumes.
# Override by mounting your own files at /etc/hwexp/ or /var/lib/hwexp/.
RUN mkdir -p /etc/hwexp/conf.d /etc/hwexp/custom.d /var/lib/hwexp
COPY configs/hwexp.yaml /etc/hwexp/hwexp.yaml
COPY configs/mappings.yaml /etc/hwexp/mappings.yaml
EXPOSE 9200
ENTRYPOINT ["/usr/bin/hwexp"]
CMD ["--config", "/etc/hwexp/hwexp.yaml"]
