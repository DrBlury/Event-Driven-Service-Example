#!/bin/bash

set -euo pipefail

FILE="/pods/main"

if [[ -f "$FILE" && "${PERSISTENCE:-0}" == 1 ]]; then
  echo "Restoring from pod"

  /opt/code/localstack/bin/localstack-cli pod load file://${FILE}
else
  echo "No pod state found"
fi
