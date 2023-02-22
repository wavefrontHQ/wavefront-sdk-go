#!/usr/bin/env bash
set -eo pipefail

echo "creating an scratch go module and installing the latest sdk release there"
echo "this causes http://pkg.go.dev to show this release as the latest version"
echo "NOTE: version updates on pkg.go.dev are not immediate. Expect a 15m+ delay."

if [ -z "$RELEASE_VERSION" ]; then
  echo "checking github for latest release version"
  echo "  (you can also set the RELEASE_VERSION variable to skip this step)"
  RELEASE_VERSION="$(gh release list --limit 1 | tail -n 1 | awk '{ print $1 }')"
fi

echo "using release version: $RELEASE_VERSION"

set -eou pipefail
cd "$(mktemp -d)"
go mod init example.com/scratch
go get "github.com/wavefronthq/wavefront-sdk-go@$RELEASE_VERSION"

echo "Done. Changes should show up within 15m"
echo "  at https://pkg.go.dev/github.com/wavefronthq/wavefront-sdk-go@v$RELEASE_VERSION"
