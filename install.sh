#!/usr/bin/env bash

set -euo pipefail || exit 1

log() {
	echo "$@" >&2
}

if [[ $HOSTTYPE != "x86_64" ]]; then
	log "arch $HOSTTYPE is not supported"
	log "please see https://sail.dev/docs/installation"
	exit 1
fi

if ! command -v curl > /dev/null && ! command -v wget > /dev/null; then
	log "please install curl or wget to use this script"
	exit 1
fi

download() {
	if command -v curl > /dev/null; then
		curl --progress-bar -L "$1"
	elif command -v wget > /dev/null; then
		wget "$1" -O -
	fi
}

latestReleaseURL() {
	log "finding latest release"
	local os=$1
	download https://api.github.com/repos/cdr/sail/releases/latest |
		jq -r ".assets[]
		| select(.name | test(\"sail-${os}-amd64.tar\"))
		| .browser_download_url"
}

downloadArchive() {
	local os=$1
	local downloadURL

	downloadURL="$(latestReleaseURL "$os")"

	log "downloading archive"

	download "$downloadURL"
}

install() {
	local os=$1
	local archive
	archive=$(mktemp)

	log "ensuring /usr/local/bin"
	sudo mkdir -p /usr/local/bin

	downloadArchive "$os" > "$archive"

	log "extracting archive into /usr/local/bin"
	sudo tar -xf "$archive" -C /usr/local/bin
}

case $OSTYPE in
linux-gnu*)
	install linux
	;;
darwin*)
	install darwin
	;;
*)
	log "$OSTYPE is not supported at the moment for automatic installation"
	log "please see https://sail.dev/docs/installation"
	exit 1
	;;
esac

log "sail has been installed into /usr/local/bin/sail"
# shellcheck disable=SC2016
log 'please ensure /usr/local/bin is in your $PATH'
