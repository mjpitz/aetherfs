#!/usr/bin/env bash
# Copyright (C) The AetherFS Authors - All Rights Reserved
# See LICENSE for more information.

set -e -o pipefail

golangci-lint run --fix

cd web

if [[ -e node_modules/.cache/platform.txt ]]; then
  last_platform="$(cat node_modules/.cache/platform.txt)"
fi

package_json_last_modified=$(date +%s -r package.json)
node_modules_last_modified=$(date +%s -r node_modules || echo -n "")

# reinstall dependencies if there are new dependencies or if we switch platforms
if [[ $(( package_json_last_modified )) -gt $(( node_modules_last_modified )) ]] || [[ "${last_platform}" != "${current_platform}" ]]; then
  npm install
  mkdir -p node_modules/.cache
  echo -n "${current_platform}" > node_modules/.cache/platform.txt
  npm audit fix
fi

npm run lint
