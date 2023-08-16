#!/usr/bin/env bash
set -efuo pipefail

which fly || (
  echo "This requires fly to be installed"
  echo "Download the binary from https://github.com/concourse/concourse/releases or from the Runway Concourse: https://runway-ci.eng.vmware.com"
  exit 1
)

fly -t runway sync || (
  echo "This requires the runway target to be set"
  echo "Create this target by running 'fly -t runway login -c https://runway-ci.eng.vmware.com -n tobs-k8s-group'"
  exit 1
)

pipeline_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
VERSION=${VERSION:-"0.14.0"}
OSM_ENVIRONMENT=${OSM_ENVIRONMENT:-"production"}
echo "using OSM_ENVIRONMENT: ${OSM_ENVIRONMENT}. Valid environments are beta and production"

fly --target runway set-pipeline \
    --pipeline "osspi-wavefront-sdk-go-${VERSION}" \
    --config "${pipeline_dir}/pipeline.yml" \
    --var osm-environment="${OSM_ENVIRONMENT}" \
    --var version="${VERSION}"

fly -t runway order-pipelines --alphabetical > /dev/null