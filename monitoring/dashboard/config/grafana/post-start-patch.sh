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

# 2. Inject CSS link into index.html
echo "Injecting CSS into index.html..."
if grep -q "custom-grafana.css" "$INDEX_HTML"; then
    echo "CSS link already present in index.html"
else
    # Replace </head> with our CSS link immediately before it.
    # Inline replacement avoids \n in sed replacement (not reliable in BusyBox/Alpine).
    sed -i 's|</head>|<link rel="stylesheet" href="/public/img/custom-grafana.css"></head>|' "$INDEX_HTML" || echo "Warning: failed to inject CSS"
    echo "CSS injected."
fi

# 3. Strip "Powered by Grafana" literal from compiled JS bundles
echo "Stripping text from JS bundles..."
find /usr/share/grafana/public/build/ -name "*.js" -type f -exec sed -i 's/Powered by Grafana//g' {} + || echo "Warning: failed to strip JS"

# 4. Write JS kiosk fix.
# Uses inline style setProperty(...,'important') which wins over all CSS rules including Grafana's
# dynamic CSS-in-JS injections (which override our stylesheet !important rules by coming later).
echo "Writing kiosk JS fix..."
JS_DEST_FILE="$CSS_DEST_DIR/kiosk-fix.js"
cat > "$JS_DEST_FILE" << 'JSEOF'
(function(){
  function fixKiosk(){
    // 1. Force body overflow off — Grafana dynamically injects body{overflow-y:auto}
    //    Inline style setProperty wins over all CSS regardless of !important cascade order.
    document.body.style.setProperty('overflow','hidden','important');
    document.body.style.setProperty('overflow-y','hidden','important');

    // 2. Clip the .main-view wrapper (and its children) that extend 33px past the viewport.
    //    The react-grid-layout content IS fully in viewport (verified via CDP); the extra
    //    33px is empty flex space below the last panel row — safe to clip.
    var selectors=['.main-view','[class*="page-content"]','[class*="page-panes"]'];
    for(var s=0;s<selectors.length;s++){
      var els=document.querySelectorAll(selectors[s]);
      for(var i=0;i<els.length;i++){
        els[i].style.setProperty('overflow','hidden','important');
        els[i].style.setProperty('max-height','100vh','important');
      }
    }

    // 3. Hide baron custom scrollbar divs (baron JS sets inline display:block after render)
    var baron=document.querySelectorAll('.baron__bar,.baron__track');
    for(var i=0;i<baron.length;i++){baron[i].style.setProperty('display','none','important');}
  }

  var mo=new MutationObserver(fixKiosk);
  mo.observe(document.documentElement,{childList:true,subtree:true,attributes:true});
  setInterval(fixKiosk,250);
  // Also run immediately in case DOM is already ready
  if(document.readyState!=='loading'){fixKiosk();}
  else{document.addEventListener('DOMContentLoaded',fixKiosk);}
})();
JSEOF
echo "JS written."

# 5. Inject JS script tag before </body>
echo "Injecting JS into index.html..."
if grep -q "kiosk-fix.js" "$INDEX_HTML"; then
    echo "JS already present in index.html"
else
    sed -i 's|</body>|<script src="/public/img/kiosk-fix.js"></script></body>|' "$INDEX_HTML" || echo "Warning: failed to inject JS"
    echo "JS injected."
fi

touch "$STAMP_FILE"
echo "Grafana UI patch script finished."
exit 0
