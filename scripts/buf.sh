#!/usr/bin/env bash
# Copyright (C) The AetherFS Authors - All Rights Reserved
# See LICENSE for more information.

set -e -o pipefail

rm -rf api gen
buf lint
buf generate
