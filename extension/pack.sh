#!/usr/bin/env bash

set -e

cd $(dirname "$0")

VERSION=$(jq -r ".version" ./out/manifest.json)
SRC_DIR="./out"
OUTPUT_DIR="./packed-extensions"

mkdir -p "$OUTPUT_DIR"

# Firefox extension (done first because web-ext verifies manifest)
if [ -z "$WEB_EXT_API_KEY" ]; then
	web-ext build --source-dir="$SRC_DIR" --artifacts-dir="$OUTPUT_DIR" --overwrite-dest
	mv "$OUTPUT_DIR/sail-$VERSION.zip" "$OUTPUT_DIR/sail-$VERSION.firefox.zip"
else
	# Requires $WEB_EXT_API_KEY and $WEB_EXT_API_SECRET from addons.mozilla.org.
	web-ext sign --source-dir="$SRC_DIR" --artifacts-dir="$OUTPUT_DIR" --overwrite-dest
	mv "$OUTPUT_DIR/sail-$VERSION.xpi" "$OUTPUT_DIR/sail-$VERSION.firefox.xpi"
fi

# Chrome extension
rm "$OUTPUT_DIR/sail-$VERSION.chrome.zip" || true
zip -j "$OUTPUT_DIR/sail-$VERSION.chrome.zip" "$SRC_DIR"/*
