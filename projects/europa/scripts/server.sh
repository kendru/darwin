#!/usr/bin/env bash

root_dir="$(dirname ${BASH_SOURCE[0]})/.."
port="${PORT:-8080}"

python3 \
  -m http.server \
  --bind 127.0.0.1 \
  --directory "${root_dir}/public" \
  "$port"
