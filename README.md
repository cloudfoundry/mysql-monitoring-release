# mysql-monitoring-release

## Adding the UAA client

A client is required for sending notifications. With the `cf-uaac` gem installed:

```
$ uaac target https://uaa.${YOUR_SYSTEM_DOMAIN}
# Enter the secret from 'Credentials -> UAA -> Admin Client Credentials' when prompted
$ uaac token client get admin
$ uaac client add mysql-monitoring \
  --authorized_grant_types client_credentials \
  --authorities emails.write
```

## Setup

In your bosh deployment manifest make sure you:

1. Add the value of `mysql-monitoring.replication-canary.canary_username` to `cf_mysql.mysql.server_audit_excluded_users`
1. Add the value of `mysql-monitoring.replication-canary.canary_username` to `cf_mysql.broker.quota_enforcer.ignored_users`
1. Add the following to `cf_mysql.mysql.seeded_databases`:
```
- name: VALUE_OF_mysql-monitoring.replication-canary.canary_database
  username: VALUE_OF_mysql-monitoring.replication-canary.canary_username
  password: VALUE_OF_mysql-monitoring.replication-canary.canary_password
```
