#!/bin/bash
# Stop and disable the service before removal
if command -v systemctl >/dev/null 2>&1; then
    systemctl stop hwexp.service || true
    systemctl disable hwexp.service || true
else
    echo "systemctl not found, skipping service disablement."
fi

# Close firewall port for hwexp metrics endpoint (9200/tcp)
HWEXP_PORT=9200
if command -v firewall-cmd >/dev/null 2>&1 && firewall-cmd --state >/dev/null 2>&1; then
    firewall-cmd --zone=public --remove-port=${HWEXP_PORT}/tcp --permanent || true
    firewall-cmd --reload || true
    echo "firewalld: closed port ${HWEXP_PORT}/tcp"
elif command -v ufw >/dev/null 2>&1 && ufw status | grep -q "Status: active"; then
    ufw delete allow ${HWEXP_PORT}/tcp || true
    echo "ufw: closed port ${HWEXP_PORT}/tcp"
fi
