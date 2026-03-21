#!/bin/sh
# Grafana UI Post-Start Patch
# Hides "Powered by Grafana" and footer elements for kiosk displays.

echo "Starting Grafana UI patch script..."

STAMP_FILE="/usr/share/grafana/public/img/.patch-applied"
INDEX_HTML="/usr/share/grafana/public/views/index.html"
CSS_DEST_DIR="/usr/share/grafana/public/img"
CSS_DEST_FILE="$CSS_DEST_DIR/custom-grafana.css"
CSS_SRC_FILE="/opt/grafana-patch/custom-grafana.css"

# Always apply the patch on startup just to be safe, but only inject once
echo "Applying Grafana UI patch..."

# 1. Copy CSS to a directory that Grafana actually serves
echo "Copying CSS..."
cp "$CSS_SRC_FILE" "$CSS_DEST_FILE" || echo "Warning: failed to copy CSS"

# 2. Inject CSS link into index.html just before </head>
echo "Injecting CSS into index.html..."
if grep -q "custom-grafana.css" "$INDEX_HTML"; then
    echo "CSS link already present in index.html"
else
    # We use a relative path here so it works regardless of root_url
    sed -i 's|</head>|<link rel="stylesheet" href="public/img/custom-grafana.css">\n</head>|' "$INDEX_HTML" || echo "Warning: failed to inject CSS"
    echo "CSS injected."
fi

# 3. Strip "Powered by Grafana" literal from compiled JS bundles
echo "Stripping text from JS bundles..."
find /usr/share/grafana/public/build/ -name "*.js" -type f -exec sed -i 's/Powered by Grafana//g' {} + || echo "Warning: failed to strip JS"

touch "$STAMP_FILE"
echo "Grafana UI patch script finished."
exit 0
