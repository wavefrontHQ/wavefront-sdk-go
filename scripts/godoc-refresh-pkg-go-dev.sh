#!/usr/bin/env bash
set -eou pipefail

cd "$(mktemp -d)"
go mod init example.com/scratch
go get "github.com/wavefronthq/wavefront-sdk-go@$RELEASE_VERSION"