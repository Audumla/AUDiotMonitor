#!/bin/bash
# Enable and start the service after installation
if command -v systemctl >/dev/null 2>&1; then
    systemctl daemon-reload || true
    systemctl enable hwexp.service || true
    systemctl start hwexp.service || true
else
    echo "systemctl not found, skipping service enablement."
fi
