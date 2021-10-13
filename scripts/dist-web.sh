#!/usr/bin/env bash
set -e -o pipefail

cd internal/web

current_platform="$(uname -s)-$(uname -m)"
last_platform=""

if [[ -e node_modules/.cache/platform.txt ]]; then
  last_platform="$(cat node_modules/.cache/platform.txt)"
fi

package_json_last_modified=$(date +%s -r package.json)
node_modules_last_modified=$(date +%s -r node_modules)

# reinstall dependencies if there are new dependencies or if we switch platforms
if [[ $(( package_json_last_modified )) -gt $(( node_modules_last_modified )) ]] || [[ "${last_platform}" != "${current_platform}" ]]; then
  npm install
  echo -n "${current_platform}" > node_modules/.cache/platform.txt
  npm audit fix
fi

npm run build
