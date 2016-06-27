#!/bin/bash

set -eux

set_env() {
  MY_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
  CI_DIR="$( cd ${MY_DIR}/.. && pwd )"
  WORKSPACE_DIR="$( cd ${CI_DIR}/.. && pwd )"

  CF_MYSQL_CI="${WORKSPACE_DIR}/cf-mysql-ci}"
  NOTIFICATIONS_RELEASE_DIR="${WORKSPACE_DIR}/notifications-release-repo"
  DEPLOYMENTS_CORE_SERVICES_DIR="${WORKSPACE_DIR}/deployments-core-services"

  : "${OUTPUT_FILE:?}"
  : "${ENV_TARGET_FILE:?}"
  : "${ENV_METADATA:?}"

  # If the output file is not absolute, assume it is relative to the WORKSPACE_DIR
  if [[ "${OUTPUT_FILE}" != /* ]]; then
    OUTPUT_FILE="${WORKSPACE_DIR}/${OUTPUT_FILE}"
  fi

  source "${CF_MYSQL_CI}/scripts/utils.sh"

  pushd "${WORKSPACE_DIR}" > /dev/null
    DIRECTOR_IP=$(cat "${ENV_TARGET_FILE}")

    BOSH_USERNAME="$(jq_val "bosh_user" "${ENV_METADATA}")"
    BOSH_PASSWORD="$(jq_val "bosh_password" "${ENV_METADATA}")"

    bosh -t "${DIRECTOR_IP}" login "${BOSH_USERNAME}" "${BOSH_PASSWORD}"
    export DIRECTOR_UUID=$(bosh status --uuid)
  popd > /dev/null
}

build_manifest() {
  pushd "${NOTIFICATIONS_RELEASE_DIR}" > /dev/null

    DB_PROPERTIES="${NOTIFICATIONS_RELEASE_DIR}/bosh-lite/notifications-db-stub.yml"
    cat <<DB_PROPERTIES > ${DB_PROPERTIES}
properties:
  notifications:
    database:
      url: tcp://root:password@10.244.7.2:3306/notifications_db
DB_PROPERTIES

    $RELEASE_DIR/generate_deployment_manifest \
        warden \
        bosh-lite/notifications-stub.yml \
        "${DEPLOYMENTS_CORE_SERVICES_DIR}/bosh-lite-stubs/notifications/smtp-stub.yml" \
        bosh-lite/notifications-db-stub.yml > bosh-lite/manifests/notifications-manifest.yml

      perl -pi -e "s/PLACEHOLDER-DIRECTOR-UUID/$DIRECTOR_UUID/g" bosh-lite/manifests/notifications-manifest.yml

    cp bosh-lite/manifests/notifications-manifest.yml ${OUTPUT_FILE}
  popd > /dev/null
}

main() {
  set_env
  build_manifest
}

main
