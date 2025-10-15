#!/bin/bash

set -euo pipefail

if ! test -f /opt/code/localstack/bin/localstack-cli; then
  echo "Installing global localstack cli"

  tmp_dir=$(mktemp -d)

  curl -s -L https://github.com/localstack/localstack-cli/releases/download/v2.2.0/localstack-cli-2.2.0-linux-amd64-onefile.tar.gz | tar xvzf - -C "$tmp_dir"
  mv "$tmp_dir/localstack" /opt/code/localstack/bin/localstack-cli

  rm -rf "$tmp_dir"
fi
