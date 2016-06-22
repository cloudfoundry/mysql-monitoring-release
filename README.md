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
