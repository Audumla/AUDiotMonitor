#!/bin/bash
# Enable and start the service after installation

# Set identity.host to the actual hostname on first install (if still at default)
if [ -f /etc/hwexp/hwexp.yaml ]; then
    REAL_HOST=$(hostname -f 2>/dev/null || hostname 2>/dev/null || echo "localhost")
    if grep -q 'host: "localhost"' /etc/hwexp/hwexp.yaml; then
        sed -i "s|host: \"localhost\"|host: \"$REAL_HOST\"|" /etc/hwexp/hwexp.yaml
    fi
fi

if command -v systemctl >/dev/null 2>&1; then
    systemctl daemon-reload || true
    systemctl enable hwexp.service || true
    systemctl start hwexp.service || true
else
    echo "systemctl not found, skipping service enablement."
fi

# Open firewall port for hwexp metrics endpoint (9200/tcp)
HWEXP_PORT=9200
if command -v firewall-cmd >/dev/null 2>&1 && firewall-cmd --state >/dev/null 2>&1; then
    firewall-cmd --zone=public --add-port=${HWEXP_PORT}/tcp --permanent || true
    firewall-cmd --reload || true
    echo "firewalld: opened port ${HWEXP_PORT}/tcp"
elif command -v ufw >/dev/null 2>&1 && ufw status | grep -q "Status: active"; then
    ufw allow ${HWEXP_PORT}/tcp || true
    echo "ufw: opened port ${HWEXP_PORT}/tcp"
else
    echo "No active firewall detected — ensure port ${HWEXP_PORT}/tcp is reachable if needed."
fi
