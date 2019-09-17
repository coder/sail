#!/bin/bash

set -e

# Firefox extension (done first because web-ext verifies manifest)
if [ -z "$AMO_JWT_ISSUER" ]; then
	web-ext build -i "node_modules/**/*" -i "src/**/*" -i "package.json" -i "tsconfig.json" -i "webpack.config.js" -i "yarn.lock"
else
	web-ext sign --api-key="$AMO_JWT_ISSUER" --api-secret="$AMO_JWT_SECRET" -i "node_modules/**/*" -i "src/**/*" -i "package.json" -i "tsconfig.json" -i "webpack.config.js" -i "yarn.lock"
fi

# Chrome extension
zip -R chrome-extension.zip manifest.json out/* logo128.png logo.svg
