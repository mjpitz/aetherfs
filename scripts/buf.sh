#!/usr/bin/env bash
set -e -o pipefail

rm -rf api gen
buf lint
buf generate
