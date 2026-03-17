#!/bin/bash
# Stop and disable the service before removal
if command -v systemctl >/dev/null 2>&1; then
    systemctl stop hwexp.service || true
    systemctl disable hwexp.service || true
else
    echo "systemctl not found, skipping service disablement."
fi
