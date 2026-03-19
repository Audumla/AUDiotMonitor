FROM grafana/grafana:latest

# Ship the AUDiot dashboard provisioning and default dashboard library inside
# the image so monitoring releases can be deployed without copying repo files.
COPY config/grafana/provisioning /etc/grafana/provisioning
COPY config/grafana/grafana.ini /etc/grafana/grafana.ini
COPY dashboards /var/lib/grafana/dashboards
COPY --chmod=0755 kiosk.sh /opt/audiot-dashboard/kiosk.sh
COPY --chmod=0755 kiosk-install.sh /opt/audiot-dashboard/kiosk-install.sh

ENV GF_AUTH_ANONYMOUS_ENABLED=true \
    GF_AUTH_ANONYMOUS_ORG_ROLE=Viewer \
    GF_SECURITY_ADMIN_PASSWORD=admin \
    GF_INSTALL_PLUGINS=yesoreyeram-infinity-datasource
