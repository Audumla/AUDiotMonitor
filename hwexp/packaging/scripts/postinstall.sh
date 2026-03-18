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
