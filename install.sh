#!/usr/bin/env bash

set -euo pipefail || exit 1

if [[ $HOSTTYPE != "x86_64" ]]; then
	log "arch $HOSTTYPE is not supported"
	log "please see https://sail.dev/docs/installation"
	exit 1
fi

downloadURL() {
	log "finding latest release"
	local os=$1
	curl --progress-bar https://api.github.com/repos/cdr/sail/releases/latest |
		jq -r ".assets[]
		| select(.name | test(\"sail-${os}-amd64.tar\"))
		| .browser_download_url"
}

log() {
	echo "$@" >&2
}

downloadArchive() {
	local os=$1
	local downloadURL

	downloadURL="$(downloadURL "$os")"

	log "downloading archive"

	if command -v curl > /dev/null; then
		curl --progress-bar -L "$downloadURL"
	elif command -v wget > /dev/null; then
		log "wget is not supported atm"
	else
		log "please install curl or wget to use this script"
		exit 1
	fi
}

install() {
	local os=$1
	local archive
	archive=$(mktemp)

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
log "please ensure /usr/local/bin is in your \$PATH"
