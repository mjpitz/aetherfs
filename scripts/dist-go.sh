#!/usr/bin/env bash
# Copyright (C) The AetherFS Authors - All Rights Reserved
# See LICENSE for more information.

set -e -o pipefail

go mod download
go mod verify

if [[ -z "${VERSION}" ]]; then
  goreleaser --snapshot --skip-publish --rm-dist
else
  goreleaser
fi

os=$(uname | tr '[:upper:]' '[:lower:]')
arch="$(uname -m)"
if [[ "$arch" == "x86_64" ]]; then
  ln -s "$(pwd)/dist/aetherfs_${os}_amd64/aetherfs" "$(pwd)/aetherfs"
elif [[ "$arch" == "aarch64" ]]; then
  ln -s "$(pwd)/dist/aetherfs_${os}_arm64/aetherfs" "$(pwd)/aetherfs"
fi
