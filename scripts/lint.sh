#!/usr/bin/env bash
# Copyright (C) The AetherFS Authors - All Rights Reserved
# See LICENSE for more information.

set -e -o pipefail

golangci-lint run --fix

ct lint

cd web

current_platform="$(uname -s)-$(uname -m)"
last_platform=""

if [[ -e node_modules/.cache/platform.txt ]]; then
  last_platform="$(cat node_modules/.cache/platform.txt)"
fi

function reinstall() {
  npm install
  mkdir -p node_modules/.cache
  echo -n "${current_platform}" > node_modules/.cache/platform.txt
  npm audit fix
}

# reinstall dependencies if there are new dependencies or if we switch platforms
if [[ $(( $(date +%s -r package.json) )) -gt $(( $(date +%s -r node_modules) )) ]]; then
  echo "detected new package dependencies, reinstalling dependencies"
  reinstall
elif [[ "${last_platform}" != "${current_platform}" ]]; then
  echo "changed platforms, reinstalling dependencies"
  reinstall
fi

npm run lint
