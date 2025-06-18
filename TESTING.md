# Testing mysql-monitoring Release

This document describes how to test the mysql-monitoring release, including both unit tests and system integration tests.

## Prerequisites

- Access to a BOSH environment
- `bosh` CLI installed and authenticated
- `om` CLI installed (for Ops Manager environments)
- `yq` and `jq` utilities installed
- Go 1.23+ for running unit tests

## Unit Tests

### Running Unit Tests

Unit tests can be run locally without any external dependencies:

```bash
cd src/mysql-metrics
go test ./...
```

To run specific test suites:

```bash
# Test individual packages
go test ./config ./cpu ./database_client ./disk ./diskstat ./emit ./gather ./metrics ./metrics_computer

# Run with verbose output
go test -v ./...

# Run with coverage
go test -cover ./...
```

### Test Structure

The unit tests are organized by package:
- `config/` - Configuration parsing and validation
- `cpu/` - CPU metrics collection
- `database_client/` - MySQL database connectivity
- `disk/` - Disk space metrics
- `diskstat/` - Disk I/O statistics
- `emit/` - Metrics emission and scheduling
- `gather/` - Metrics gathering coordination
- `metrics/` - Metrics processing and transmission
- `metrics_computer/` - Metrics computation logic

## System Integration Tests

System tests require a running BOSH environment with Cloud Foundry and MySQL deployed.

### Environment Setup

1. **Set up BOSH targeting**: Ensure your BOSH CLI is authenticated and targeting the correct environment.

2. **Export required environment variables**:

   This step is only required if you are targeting an Ops Manager environment such as Tanzu Platform for Cloud Foundry.
   ```bash
   # Get the CF deployment name
   export cf_deployment_name=$(bosh ds --json | yq '.Tables[].Rows[].name|select(test("^cf"))')
   
   # Set deployment configuration
   export azs='[az-0,az-1,az-2]'
   export network_name=deployment-network
   export vm_type=medium
   
   # Extract TLS certificates from Ops Manager
   export loggregator_tls_ca="$(om certificate-authority --cert-pem)"
   export loggregator_tls_client_cert="((/opsmgr/$cf_deployment_name/doppler/metron_tls_cert.cert_pem))"
   export loggregator_tls_client_key="((/opsmgr/$cf_deployment_name/doppler/metron_tls_cert.private_key_pem))"
   ```

### Deploy MySQL with Monitoring

Deploy a PXC (Percona XtraDB Cluster) instance with mysql-monitoring:

```bash
./scripts/deploy-pxc-with-monitoring -o ./operations/dev-release.yml
```

This script will:
- Deploy a PXC cluster with the mysql-monitoring release
- Configure the mysql-metrics job to collect and emit metrics
- Set up proper networking and security groups

### Configure Integration Test Environment

The `src/spec/metrics` suite uses [cf-test-helpers](github.com/cloudfoundry/cf-test-helpers) which require an integration config file to be present.

Create the integration test configuration:

```bash
jq --null-input \
    --arg api "api.sys.pcf.tasonvsphere.com" \
    --arg apps_domain "apps.pcf.tasonvsphere.com" \
    --arg "admin_user" "admin" \
    --arg "admin_password" "$(om credentials --product-name=cf --credential-reference=.uaa.admin_credentials --credential-field=password)" \
    '{
        "api": $api,
        "apps_domain": $apps_domain,
        "admin_user": $admin_user,
        "admin_password": $admin_password,
        "name_prefix": "MYSQL",
        "skip_ssl_validation": true
    }' > /tmp/integration_config.json
```

**Note**: Replace the API and apps domain values with your actual Cloud Foundry endpoints.

### Set Test Environment Variables

```bash
export CONFIG=/tmp/integration_config.json
export BOSH_DEPLOYMENT=pxc
export METRICS_SOURCE_ID=pxc-metrics-test
```

### Run Integration Tests

Run the Ginkgo test suites:

```bash
# Run all integration tests
cd src/spec
ginkgo -r

# Run specific test suite
ginkgo metrics/
ginkgo mysql-diag/
ginkgo templates/

# Run with verbose output
ginkgo -v -r

# Run tests in parallel
ginkgo -p -r
```

## Test Suites Description

### `src/spec/metrics/`
Tests the mysql-metrics functionality:
- Metrics collection from MySQL databases
- Metrics emission to loggregator
- Configuration validation
- Error handling and recovery

### `src/spec/mysql-diag/`
Tests the mysql-diag diagnostic tool:
- Database connectivity diagnostics
- Disk space analysis
- Galera cluster health checks
- Agent communication

### `src/spec/templates/`
Tests BOSH job templates:
- Template rendering with various configurations
- Configuration file generation
- Service startup scripts

## Troubleshooting

### Common Issues

1. **BOSH authentication errors**: Ensure you're logged into BOSH and targeting the correct environment:
   ```bash
   bosh env
   bosh login
   ```

2. **Missing environment variables**: Verify all required environment variables are set:
   ```bash
   env | grep -E "(cf_deployment_name|azs|network_name|vm_type|loggregator_)"
   ```

3. **Certificate issues**: If TLS certificate extraction fails, verify Ops Manager connectivity:
   ```bash
   om certificate-authority --cert-pem
   ```

4. **Test failures**: Check BOSH deployment status and logs:
   ```bash
   bosh deployments
   bosh vms -d pxc
   bosh logs -d pxc mysql-metrics
   ```

### Log Analysis

Monitor mysql-metrics logs during testing:

```bash
# Follow logs in real-time
bosh logs -d pxc mysql-metrics -f

# Download logs for analysis
bosh logs -d pxc mysql-metrics
```

The logs will show:
- Metrics collection activity
- Database connection status
- Loggregator emission success/failures
- Configuration parsing results

## Development Workflow

For development and testing iterations:

1. Make code changes
2. Run unit tests: `go test ./...`
3. Create dev release: `./scripts/create-and-upload-dev-release`
4. Redeploy with monitoring: `./scripts/deploy-pxc-with-monitoring -o ./operations/dev-release.yml`
5. Run integration tests: `cd src/spec && ginkgo -r`

This workflow ensures both unit-level correctness and system-level integration. 