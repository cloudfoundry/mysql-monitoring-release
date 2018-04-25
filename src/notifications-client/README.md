# notifications-client

## Description

This client library connects to a [notifications service](https://github.com/cloudfoundry-incubator/notifications), using credentials for a UAA client which has the appropriate notifications-related scopes. It provides an interface for sending emails using the underlying notifications service.

## Testing
To run the integration tests
- You need a Cloud Foundry with a deployed notifications service (ERT provides this)
- Ensure that you also set the following environment variables. Some defaults are provided (with valid values for bosh-lite):
  - NOTIFICATIONS_MAILINATOR_TOKEN
  - NOTIFICATIONS_UAA_DOMAIN
  - NOTIFICATIONS_NOTIFICATIONS_DOMAIN
  - NOTIFICATIONS_RECIPIENT_EMAIL
  - NOTIFICATIONS_CLIENT_SECRET
  - NOTIFICATIONS_CLIENT_USERNAME
  - NOTIFICATIONS_UAA_ADMIN_CLIENT_USERNAME
  - NOTIFICATIONS_UAA_ADMIN_CLIENT_SECRET
- run `bin/test-integration`
- You can find an example concourse task for running these tests [here](https://github.com/pivotal-cf-experimental/mysql-monitoring-ci/blob/ad48b14983eedb428bee3e4ba86edff3475c9a1f/ci/scripts/test-integration) 
