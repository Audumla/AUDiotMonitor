#!/bin/bash
# AUDiot Kiosk Starter
# Recommended for use in .xsession or systemd user service

URL="http://127.0.0.1:3000/d/hwexp-panel-main/audiot-panel-display?kiosk"

# Hide cursor and disable screensaver
xset s off
xset -dpms
xset s noblank

/usr/bin/chromium-browser --kiosk --app="$URL" --window-size=1920,440 --window-position=0,0
