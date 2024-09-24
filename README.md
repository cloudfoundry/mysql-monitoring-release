# mysql-monitoring-release

## Setup

### Deploying with MySQL Metrics and a PXC release mysql deployment
1. Add the [operations/pxc-add-metrics.yml](https://github.com/cloudfoundry-incubator/mysql-monitoring-release/blob/master/operations/pxc-add-metrics.yml) ops file to your [pxc deployment](https://github.com/cloudfoundry-incubator/pxc-release).
2. Change the [operations/loggregator_vars_template.yml](https://github.com/cloudfoundry-incubator/mysql-monitoring-release/blob/master/operations/loggregator_vars_template.yml) to have the correct name of your director and cf deployment, so that the loggregator agent gets the cert from the loggregator deployment in cf, and add this as a vars file.
3. Set the `loggregator_agent_deployment` variable for the loggregator agent job to tag your metrics with the deployment name.

## Deploying as the backing store for Cloud Foundry

In order to use [cf-mysql-release](https://github.com/cloudfoundry/cf-mysql-release) as the internal database for Cloud Foundry,
various components of Cloud Foundry like UAA and CAPI require the backing database to be online before they will start successfully.
However, the replication canary also has a dependency on UAA to obtain a token for sending emails via the notifications service.

The easiest way to break this dependency cycle when adding this monitoring to a Cloud Foundry deployment is as follows:
- ensure `max_in_flight=1` (mysql always needs this anyway)
- ensure `serial: true` in the update block
- re-order jobs in manifest as follows:
 - consul
 - other jobs e.g. NATS, router
 - mysql DBs
 - mysql proxies
 - UAA/CC
 - replication-canary
 - everything else including Diego
