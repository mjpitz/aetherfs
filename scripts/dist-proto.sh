#!/usr/bin/env bash
# Copyright (C) The AetherFS Authors - All Rights Reserved
# See LICENSE for more information.


readonly lock_file="dist/aetherfs_proto.lock"

if { set -C; true >${lock_file}; }; then
  echo "[lock] obtained"
  mkdir -p dist/aetherfs_proto
  cp -R proto/* dist/aetherfs_proto

  echo "[tar] packaging proto files"
  tar -czf dist/aetherfs_proto.tar.gz -C dist/aetherfs_proto/ .
else
  echo "[lock] exists, skipping"
  exit
fi
