#!/usr/bin/env bash
set -eou pipefail

which godoc >/dev/null || (
  cat <<-'EOF'
The godoc cli is not installed.
Install godoc somewhere on your $PATH.

For example, if /usr/local/bin is on your $PATH:
  GOBIN=/usr/local/bin/ go install golang.org/x/tools/cmd/godoc@latest
EOF
)
