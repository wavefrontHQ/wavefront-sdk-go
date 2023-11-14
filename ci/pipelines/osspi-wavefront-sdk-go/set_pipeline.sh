#!/usr/bin/env bash
set -efuo pipefail

which fly || (
  echo "This requires fly to be installed"
  echo "Download the binary from https://github.com/concourse/concourse/releases or from the Runway Concourse: https://runway-ci-sfo.eng.vmware.com"
  exit 1
)

TARGET=runway-sfo
fly -t runway sync || (
  echo "This requires the runway target to be set"
  echo "Create this target by running 'fly -t "${TARGET}" login -c https://runway-ci-sfo.eng.vmware.com -n cpo-team-troll'"
  exit 1
)

pipeline_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
VERSION=${VERSION:-"Latest"}
OSM_ENVIRONMENT=${OSM_ENVIRONMENT:-"production"}
echo "using OSM_ENVIRONMENT: ${OSM_ENVIRONMENT}. Valid environments are beta and production"

fly --target "${TARGET}" set-pipeline \
    --pipeline "wavefront-sdk-go-osspi" \
    --config "${pipeline_dir}/pipeline.yml" \
    --var osm-environment="${OSM_ENVIRONMENT}" \
    --var version="${VERSION}"

fly -t "${TARGET}" order-pipelines --alphabetical > /dev/null