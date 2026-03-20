FROM grafana/grafana:latest

# Ship the AUDiot dashboard provisioning and default dashboard library inside
# the image so monitoring releases can be deployed without copying repo files.
COPY config/grafana/provisioning /etc/grafana/provisioning
COPY config/grafana/grafana.ini /etc/grafana/grafana.ini
COPY dashboards /opt/audiot-dashboard/dashboards
COPY examples /opt/audiot-dashboard/examples
COPY --chmod=0755 set-dashboard.sh /opt/audiot-dashboard/set-dashboard.sh
COPY --chmod=0755 entrypoint.sh /opt/audiot-dashboard/entrypoint.sh
COPY --chmod=0755 kiosk.sh /opt/audiot-dashboard/kiosk.sh
COPY --chmod=0755 kiosk-install.sh /opt/audiot-dashboard/kiosk-install.sh

USER root
RUN chown -R 472:472 /etc/grafana/provisioning /etc/grafana/grafana.ini /opt/audiot-dashboard
USER grafana

ENV GF_AUTH_ANONYMOUS_ENABLED=true \
    GF_AUTH_ANONYMOUS_ORG_ROLE=Viewer \
    GF_SECURITY_ADMIN_PASSWORD=admin \
    GF_INSTALL_PLUGINS=yesoreyeram-infinity-datasource

ENTRYPOINT ["/opt/audiot-dashboard/entrypoint.sh"]
