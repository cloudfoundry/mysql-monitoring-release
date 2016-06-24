#!/bin/bash

set -eu

MY_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
RELEASE_DIR="$( cd "${MY_DIR}/.." && pwd )"
CF_MYSQL_RELEASE_DIR="$( cd "${RELEASE_DIR}/../cf-mysql-release" && pwd )"

DIRECTOR_IP=${DIRECTOR_IP:-192.168.50.4}
BOSH_LITE_USERNAME=${BOSH_LITE_USERNAME:-admin}
BOSH_LITE_PASSWORD=${BOSH_LITE_PASSWORD:-admin}

tmpdir=$(mktemp -d /tmp/mysql_manifest.XXXXX)
trap '{ rm -rf ${tmpdir}; }' EXIT

pushd "${RELEASE_DIR}" > /dev/null

  bosh -n target "${DIRECTOR_IP}"
  bosh -n login "${BOSH_LITE_USERNAME}" "${BOSH_LITE_PASSWORD}"
  DIRECTOR_UUID="$(bosh status --uuid)"
  sed "s/REPLACE_WITH_DIRECTOR_UUID/$DIRECTOR_UUID/g" \
    "${CF_MYSQL_RELEASE_DIR}/manifest-generation/bosh-lite-stubs/cf-manifest.yml" \
    > "${tmpdir}/bosh-lite-cf-manifest.yml"

  ${CF_MYSQL_RELEASE_DIR}/scripts/generate-deployment-manifest \
    -c "${tmpdir}/bosh-lite-cf-manifest.yml" \
    -p ${RELEASE_DIR}/manifest-generation/bosh-lite-stubs/property-overrides-monitoring.yml \
    -p ${CF_MYSQL_RELEASE_DIR}/manifest-generation/bosh-lite-stubs/property-overrides.yml \
    -i ${RELEASE_DIR}/manifest-generation/bosh-lite-stubs/iaas-settings-monitoring.yml \
    -i ${CF_MYSQL_RELEASE_DIR}/manifest-generation/bosh-lite-stubs/iaas-settings.yml \
    -j ${RELEASE_DIR}/manifest-generation/examples/job-overrides.yml \
    > mysql-monitoring.yml

  bosh deployment mysql-monitoring.yml
popd > /dev/null

echo "MySQL monitoring manifest was generated at ${RELEASE_DIR}/mysql-monitoring.yml"
