#!/usr/bin/env bash
set -eou pipefail

which golangci-lint >/dev/null || (
  cat <<-'EOF'
The golangci-lint cli is not installed.

For install instructions, see the documentation here:
https://golangci-lint.run/usage/install/#local-installation

EOF
)
