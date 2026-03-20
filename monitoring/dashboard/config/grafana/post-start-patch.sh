#!/bin/sh
# Grafana UI Post-Start Patch
# Hides "Powered by Grafana" and footer elements for kiosk displays.

set -eu

STAMP_FILE="/usr/share/grafana/public/custom/.patch-applied"
INDEX_HTML="/usr/share/grafana/public/views/index.html"
CSS_DEST_DIR="/usr/share/grafana/public/custom"
CSS_DEST_FILE="$CSS_DEST_DIR/custom-grafana.css"
CSS_SRC_FILE="/opt/grafana-patch/custom-grafana.css"

if [ -f "$STAMP_FILE" ]; then
    echo "Patch already applied. Skipping."
    exit 0
fi

echo "Applying Grafana UI patch..."

# 1. Create custom directory and copy CSS
mkdir -p "$CSS_DEST_DIR"
cp "$CSS_SRC_FILE" "$CSS_DEST_FILE"

# 2. Inject CSS link into index.html just before </head>
if ! grep -q "custom-grafana.css" "$INDEX_HTML"; then
    sed -i 's|</head>|<link rel="stylesheet" href="public/custom/custom-grafana.css">\n</head>|' "$INDEX_HTML"
fi

# 3. Strip "Powered by Grafana" literal from compiled JS bundles
# Replaces it with an empty string so it doesn't render in the UI
find /usr/share/grafana/public/build/ -name "*.js" -type f -exec sed -i 's/Powered by Grafana//g' {} +

# Mark as patched
touch "$STAMP_FILE"
echo "Grafana UI patch applied successfully."
