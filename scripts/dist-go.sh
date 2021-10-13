#!/usr/bin/env bash
set -e -o pipefail

go mod download
go mod verify

goreleaser --snapshot --skip-publish --rm-dist
