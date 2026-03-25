# audiot-kiosk-tools
#
# Thin image containing pre-built kiosk host binaries.
# Used by the DietPi bootstrap installer to extract binaries without requiring
# a Rust/C toolchain on the target device:
#
#   docker run --rm audumla/audiot-kiosk-tools \
#       cat /usr/local/bin/wl-gammarelay > /usr/local/bin/wl-gammarelay
#   chmod +x /usr/local/bin/wl-gammarelay
#
# Targets: linux/arm64 (Raspberry Pi / DietPi), linux/amd64

# ── Stage 1: wl-gammarelay ────────────────────────────────────────────────────
FROM rust:1-slim-bookworm AS wl-gammarelay-builder
RUN apt-get update && apt-get install -y --no-install-recommends \
        libdbus-1-dev \
        pkg-config \
    && rm -rf /var/lib/apt/lists/*
RUN cargo install wl-gammarelay --locked

# ── Stage 2: minimal runtime image ───────────────────────────────────────────
FROM debian:bookworm-slim
LABEL org.opencontainers.image.title="audiot-kiosk-tools" \
      org.opencontainers.image.description="Pre-built kiosk host binaries (wl-gammarelay)" \
      org.opencontainers.image.source="https://github.com/audumla/AUDiotMonitor"

RUN apt-get update && apt-get install -y --no-install-recommends \
        libdbus-1-3 \
    && rm -rf /var/lib/apt/lists/*

COPY --from=wl-gammarelay-builder /usr/local/cargo/bin/wl-gammarelay /usr/local/bin/wl-gammarelay

# Default: print available binaries
CMD ["ls", "-la", "/usr/local/bin/"]
