#!/bin/bash

# Path to Zsh history file
HIST_FILE="$HOME/.zsh_history"

# Check for wl-copy (for Wayland systems)
if ! command -v wl-copy &> /dev/null; then
    echo "wl-copy not found. Install it with: sudo pacman -S wl-clipboard"
    exit 1
fi

# Show reversed, deduplicated Zsh history commands in rofi and copy the selected command
SELECTED=$(tac "$HIST_FILE" | awk -F';' '{print $2}' | awk '!seen[$0]++' | rofi -dmenu -i -config ~/.config/rofi/themes/clipboard.rasi)

# If a command was selected, copy it to the clipboard using wl-copy
if [ -n "$SELECTED" ]; then
    echo -n "$SELECTED" | wl-copy
    notify-send "Copied to clipboard:" "$SELECTED"
fi

