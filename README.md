# mysql-monitoring-release

## Notifications

Requires [notifications-release](https://github.com/cloudfoundry-incubator/notifications-release)

Don't forget to first create a database for the notifications, e.g.

```
mysql -uroot -ppassword -h10.244.7.2 -e "CREATE DATABASE notifications_db"
```

Then use that in your `notifications-db-stub.yml`, e.g. `tcp://root:password@10.244.7.2:3306/notifications_db`

## Adding the UAA client

A client is required for sending notifications. With the `cf-uaac` gem installed:

```
$ uaac target https://uaa.${YOUR_SYSTEM_DOMAIN}
# Enter the secret from 'Credentials -> UAA -> Admin Client Credentials' when prompted
$ uaac token client get admin
$ uaac client add mysql-monitoring \
  --authorized_grant_types client_credentials \
  --authorities notifications.write,critical_notifications.write,emails.write,emails.write \
  --secret ${MYSQL_MONITORING_NOTIFICATIONS_CLIENT_SECRET:-"REPLACE_WITH_CLIENT_SECRET"}
```

## Setup
### Deploying with the replication canary
In your bosh deployment manifest make sure you:

1. Add the value of `mysql-monitoring.replication-canary.canary_username` to `cf_mysql.mysql.server_audit_excluded_users`
1. Add the value of `mysql-monitoring.replication-canary.canary_username` to `cf_mysql.broker.quota_enforcer.ignored_users`
1. Add the following to `cf_mysql.mysql.seeded_databases`:
```
- name: VALUE_OF_mysql-monitoring.replication-canary.canary_database
  username: VALUE_OF_mysql-monitoring.replication-canary.canary_username
  password: VALUE_OF_mysql-monitoring.replication-canary.canary_password
```
1. Replace the client username (`mysql-monitoring.replication-canary.notifications_client_username`) with the one created above
1. Replace the client secret (`mysql-monitoring.replication-canary.notifications_client_secret`)  with the one created above

### Deploying with MySQL Metrics and a PXC release mysql deployment
1. Add the operations/pxc-add-metrics.yml ops file to your [pxc deployment](https://github.com/cloudfoundry-incubator/pxc-release).
2. Change the operations/loggregator_vars_template.yml to have the correct name of your director and cf deployment, so that the metron agent gets the cert from the loggregator deployment in cf, and add this as a vars file.
3. Provide the `metron_agent_deployment` variable to tag your metrics with this deployment.

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
