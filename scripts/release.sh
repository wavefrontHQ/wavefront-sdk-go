#!/usr/bin/env bash
set -eo pipefail

if [ -z "$RELEASE_VERSION" ]; then
  echo "RELEASE_VERSION must be set to the new version you're trying to create"
  echo "It looks like the last tag was: $(gh release list --limit 1 | tail -n 1 | grep -o 'v[0-9.]*')"
  exit 1
fi

set -eou pipefail

git tag --annotate --message="$RELEASE_VERSION" "$RELEASE_VERSION"
git push origin "$RELEASE_VERSION"
goreleaser release -f .goreleaser.yaml

echo "review the draft and publish it if everything looks good"
echo "once it has been published, run: "
echo "> RELEASE_VERSION=$RELEASE_VERSION ./scripts/godoc-refresh-pkg-go-dev.sh"
echo "this makes pkg.go.dev aware of the new version"