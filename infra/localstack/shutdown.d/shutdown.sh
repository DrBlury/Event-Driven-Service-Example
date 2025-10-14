#!/bin/bash

set -euo pipefail

if [[ "${PERSISTENCE:-0}" == 1 ]]; then
  echo "Saving to pod"

  /opt/code/localstack/bin/localstack-cli pod save file:///pods/main
fi
