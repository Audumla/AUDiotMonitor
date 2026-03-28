# AUDiot Dashboard Image
#
# Builds on the official Grafana image and bakes in:
#   - AUDiot provisioning, configuration, and dashboard library
#   - Kiosk UI patches (CSS + JS injected into index.html at build time)
#   - wl-gammarelay host binary (extracted during install — no separate download)
#
# Build:  docker compose build
# Push:   docker push audumla/audiot-dashboard:latest
#
# Extracting kiosk tools during install:
#   docker run --rm audumla/audiot-dashboard \
#       cat /opt/audiot-dashboard/wl-gammarelay > /usr/local/bin/wl-gammarelay
#   chmod +x /usr/local/bin/wl-gammarelay

# ── Stage 0: wl-gammarelay ────────────────────────────────────────────────────
FROM rust:1-slim-bookworm AS wl-gammarelay-builder
RUN apt-get update && apt-get install -y --no-install-recommends \
        libdbus-1-dev \
        pkg-config \
    && rm -rf /var/lib/apt/lists/*
RUN cargo install wl-gammarelay-rs --locked

# ── Stage 1: kiosk-tools (standalone binary image) ────────────────────────────
# Build: docker build --target kiosk-tools -o type=local,dest=dist/ .
FROM debian:bookworm-slim AS kiosk-tools
LABEL org.opencontainers.image.title="audiot-kiosk-tools" \
      org.opencontainers.image.description="Pre-built kiosk host binaries (wl-gammarelay)" \
      org.opencontainers.image.source="https://github.com/audumla/AUDiotMonitor"
RUN apt-get update && apt-get install -y --no-install-recommends \
        libdbus-1-3 \
    && rm -rf /var/lib/apt/lists/*
COPY --from=wl-gammarelay-builder /usr/local/cargo/bin/wl-gammarelay-rs /usr/local/bin/wl-gammarelay
CMD ["ls", "-la", "/usr/local/bin/"]

# ── Stage 2: Grafana kiosk image ─────────────────────────────────────────────
FROM grafana/grafana:latest

# ── 1. Provisioning and dashboard library ─────────────────────────────────────

COPY config/grafana/provisioning /etc/grafana/provisioning
COPY config/grafana/grafana.ini  /etc/grafana/grafana.ini
COPY dashboards                  /opt/audiot-dashboard/dashboards
COPY --chmod=0755 kiosk.sh       /opt/audiot-dashboard/kiosk.sh
COPY --chmod=0755 entrypoint.sh  /opt/audiot-dashboard/entrypoint.sh

# ── 2. Kiosk UI patches ───────────────────────────────────────────────────────
#
# Copy CSS and kiosk-fix.js into the directory Grafana actually serves (/public/img/).
# Then inject a <link> and <script> into index.html so every page load gets them.
# Finally strip "Powered by Grafana" text from the compiled JS bundles.
#
# Must run as root because the grafana user has no write access to /usr/share/grafana.

COPY config/grafana/custom-grafana.css /usr/share/grafana/public/img/custom-grafana.css
COPY config/grafana/kiosk-fix.js       /usr/share/grafana/public/img/kiosk-fix.js

USER root

# Include pre-built wl-gammarelay for the host kiosk (correct arch via multi-platform build).
# install.sh/dietpi-setup.sh extract it via: docker run --rm <image> cat /opt/audiot-dashboard/wl-gammarelay
COPY --from=wl-gammarelay-builder /usr/local/cargo/bin/wl-gammarelay-rs /opt/audiot-dashboard/wl-gammarelay

RUN INDEX=/usr/share/grafana/public/views/index.html && \
    sed -i 's|</head>|<link rel="stylesheet" href="/public/img/custom-grafana.css"></head>|' "$INDEX" && \
    sed -i 's|</body>|<script src="/public/img/kiosk-fix.js"></script></body>|' "$INDEX" && \
    find /usr/share/grafana/public/build/ -name "*.js" -type f \
         -exec sed -i 's/Powered by Grafana//g' {} + && \
    chmod +x /opt/audiot-dashboard/wl-gammarelay && \
    chown -R 472:472 \
        /etc/grafana/provisioning \
        /etc/grafana/grafana.ini \
        /opt/audiot-dashboard \
        /usr/share/grafana/public/img/custom-grafana.css \
        /usr/share/grafana/public/img/kiosk-fix.js

USER grafana

# ── 3. Default environment ────────────────────────────────────────────────────

ENV GF_AUTH_ANONYMOUS_ENABLED=true \
    GF_AUTH_ANONYMOUS_ORG_ROLE=Viewer \
    GF_SECURITY_ADMIN_PASSWORD=admin \
    GF_ANALYTICS_REPORTING_ENABLED=false \
    GF_ANALYTICS_CHECK_FOR_UPDATES=false \
    GF_ANALYTICS_CHECK_FOR_PLUGIN_UPDATES=false \
    GF_INSTALL_PLUGINS=yesoreyeram-infinity-datasource

ENTRYPOINT ["/opt/audiot-dashboard/entrypoint.sh"]
