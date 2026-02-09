#!/bin/bash

WALLS_DIR="$HOME/.local/share/mywalls"
CACHE_DIR="$HOME/.cache/myone/walls"
CURRENT_WALL="$HOME/.local/share/myone/.wallpaper"

mkdir -p "$CACHE_DIR"
shopt -s nullglob

for wall in "$WALLS_DIR"/*.{png,jpg,jpeg,gif,webp}; do
    [ -f "$wall" ] || continue
    # thumb="$CACHE_DIR/$(basename "$wall").png"
    # if [ ! -f "$thumb" ]; then
    #     magick "${wall[0]}" -thumbnail 220x220^ -gravity center -extent 500x400^ "$thumb"
    # fi
    entry="$(basename "$wall")\x00icon\x1f$wall"
    options+=("$entry")
done 

selection="$(echo -e "$(printf '%s\n' "${options[@]}")" | rofi -dmenu -show-icons -i -theme "$HOME/.config/rofi/themes/wallpapers.rasi")"

echo "$selection"
if [ -n "$selection" ] && [ -f "$IMAGES_DIR/$selection" ]; then
	cp "$WALLS_DIR/$selection" "$CURRENT_WALL"
	swww img "$CURRENT_WALL" --transition-type fade --transition-duration 0.5
fi
