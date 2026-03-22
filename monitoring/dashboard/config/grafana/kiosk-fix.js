/**
 * AUDiot Kiosk Display Fix
 *
 * Injected into Grafana's index.html at image build time.
 * Fixes two issues that CSS alone cannot reliably solve:
 *
 *   1. Grafana's CSS-in-JS runtime injects `body { overflow-y: auto }` as a
 *      dynamic <style> tag after our stylesheet loads, overriding our CSS rule.
 *      We counter it with JS inline-style setProperty (wins over all CSS).
 *
 *   2. The .main-view wrapper extends ~33px past the 440px viewport (empty flex
 *      space below the last panel row). CDP measurements confirm the react-grid-
 *      layout content itself is fully in-viewport (top: 17, bottom: 427).
 *      We clip the wrappers to 100vh — nothing visible is cut off.
 *
 *   3. Grafana uses the "baron" library for custom div-based scrollbars. Baron
 *      JS sets element.style.display = 'block' after React renders, which beats
 *      plain CSS. We overwrite it with setProperty('display','none','important').
 *
 * A MutationObserver re-applies all fixes whenever the DOM changes, and a
 * 250 ms interval acts as belt-and-suspenders.
 */
(function () {
  function fixKiosk() {
    // 1. Body overflow — force both axes hidden via inline style
    document.body.style.setProperty('overflow', 'hidden', 'important');
    document.body.style.setProperty('overflow-y', 'hidden', 'important');

    // 2. Clip page-wrapper elements that extend past the viewport
    var wrappers = document.querySelectorAll(
      '.main-view, [class*="page-content"], [class*="page-panes"]'
    );
    for (var i = 0; i < wrappers.length; i++) {
      wrappers[i].style.setProperty('overflow', 'hidden', 'important');
      wrappers[i].style.setProperty('max-height', '100vh', 'important');
    }

    // 3. Hide baron custom scrollbar divs
    var baron = document.querySelectorAll('.baron__bar, .baron__track');
    for (var i = 0; i < baron.length; i++) {
      baron[i].style.setProperty('display', 'none', 'important');
    }
  }

  var mo = new MutationObserver(fixKiosk);
  mo.observe(document.documentElement, { childList: true, subtree: true, attributes: true });
  setInterval(fixKiosk, 250);

  if (document.readyState !== 'loading') {
    fixKiosk();
  } else {
    document.addEventListener('DOMContentLoaded', fixKiosk);
  }
})();
