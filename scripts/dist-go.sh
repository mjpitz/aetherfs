#!/usr/bin/env bash
# Copyright (C) The AetherFS Authors - All Rights Reserved
# See LICENSE for more information.

set -e -o pipefail

go mod download
go mod verify

goreleaser --snapshot --skip-publish --rm-dist
