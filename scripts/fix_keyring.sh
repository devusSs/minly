#!/usr/bin/env bash
set -euo pipefail

sudo systemctl restart user@1000.service
sudo apt-get update
sudo apt-get install -y gnome-keyring libsecret-tools dbus-x11
sudo killall gnome-keyring-daemon 2>/dev/null || true
eval "$(echo "" | gnome-keyring-daemon --unlock 2>/dev/null || true)"
eval "$(gnome-keyring-daemon --start 2>/dev/null)"