#!/bin/sh
# Grafana UI Post-Start Patch
# Hides "Powered by Grafana" and footer elements for kiosk displays.

echo "Starting Grafana UI patch script..."

STAMP_FILE="/usr/share/grafana/public/custom/.patch-applied"
INDEX_HTML="/usr/share/grafana/public/views/index.html"
CSS_DEST_DIR="/usr/share/grafana/public/custom"
CSS_DEST_FILE="$CSS_DEST_DIR/custom-grafana.css"
CSS_SRC_FILE="/opt/grafana-patch/custom-grafana.css"

# Always apply the patch on startup just to be safe, but only inject once
echo "Applying Grafana UI patch..."

# 1. Create custom directory and copy CSS
echo "Copying CSS..."
mkdir -p "$CSS_DEST_DIR"
cp "$CSS_SRC_FILE" "$CSS_DEST_FILE" || echo "Warning: failed to copy CSS"

# 2. Inject CSS link into index.html just before </head>
echo "Injecting CSS into index.html..."
if grep -q "custom-grafana.css" "$INDEX_HTML"; then
    echo "CSS link already present in index.html"
else
    sed -i 's|</head>|<link rel="stylesheet" href="/public/custom/custom-grafana.css">\n</head>|' "$INDEX_HTML" || echo "Warning: failed to inject CSS"
    echo "CSS injected."
fi

# 3. Strip "Powered by Grafana" literal from compiled JS bundles
echo "Stripping text from JS bundles..."
find /usr/share/grafana/public/build/ -name "*.js" -type f -exec sed -i 's/Powered by Grafana//g' {} + || echo "Warning: failed to strip JS"

touch "$STAMP_FILE"
echo "Grafana UI patch script finished."
exit 0
