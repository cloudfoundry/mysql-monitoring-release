# MySQL Metrics Testing Guide

This document describes how to run unit and integration tests for the mysql-metrics component.

## Overview

The mysql-metrics component includes both unit tests and integration tests:
- **Unit tests**: Test individual components in isolation with mocked dependencies
- **Integration tests**: Test the complete mysql-metrics binary against a real MySQL database running in Docker

## Prerequisites

### For Unit Tests Only
- **Go compiler**: Go 1.23.0 or later (toolchain 1.24.4 recommended)
- The project uses Ginkgo v2 as the testing framework

### For All Tests (Unit + Integration)
- **Go compiler**: Go 1.23.0 or later (toolchain 1.24.4 recommended)
- **Docker**: Required for running MySQL database containers during integration tests
- **Docker socket access**: The test container needs access to `/var/run/docker.sock`

## Running Tests

### All Tests (Unit + Integration)
```bash
./bin/test
```

This script:
- Runs both unit and integration tests using Docker
- Uses Ginkgo with race detection enabled
- Runs tests in parallel with 4 processes and 2 compilers
- Randomizes test execution order for better test isolation

### Unit Tests Only
```bash
./bin/test-unit
```

This script:
- Runs only unit tests (skips the `integration_test` package)
- Does not require Docker
- Uses the same Ginkgo configuration as the full test suite
- Faster execution since it doesn't spin up Docker containers

### Integration Tests Only
```bash
./bin/test-integration
```

This script:
- Runs only integration tests (tests labeled with "integration")
- Requires Docker
- Spins up a Percona MySQL 8.0 container for testing

## Test Environment Details

### Docker Environment
The integration tests use a Docker setup that:
- Builds a custom test container based on `golang:1.24`
- Installs Docker CLI inside the container
- Mounts the workspace and Docker socket
- Creates `/var/vcap/store` and `/var/vcap/data` directories for disk metrics testing
- Uses host networking mode
- Runs with privileged access (required for Docker socket access)

### Integration Test Database
Integration tests use:
- **Database**: Percona Server 8.0 (MySQL-compatible)
- **Configuration**: Empty root password, replication configured
- **Port**: Dynamically allocated and exposed
- **Setup**: Includes replication source configuration for testing replication metrics

## Test Configuration

### Environment Variables
The following environment variables are set during integration tests:
- `CGO_ENABLED=1`: Enables CGO for certain dependencies
- `TEST_VOLUME=/var/vcap/store`: Specifies test volume path for disk metrics
- `TEST_DURATION=30`: Test duration in seconds
- `MONITOR_INTERVAL=1s`: Monitoring interval for metrics collection

### Ginkgo Configuration
All test scripts use the following Ginkgo configuration:
- `--compilers=2`: Uses 2 parallel compilers
- `--procs=4`: Runs tests across 4 parallel processes
- `-race`: Enables Go race detector
- `--fail-on-pending`: Fails if there are pending tests
- `--randomize-all`: Randomizes test execution order
- `--randomize-suites`: Randomizes test suite execution order

## Test Structure

### Unit Tests
Unit tests are located throughout the codebase in `*_test.go` files alongside the source code:
- `config/` - Configuration parsing tests
- `cpu/` - CPU statistics tests
- `database_client/` - Database client tests (with mocked database)
- `disk/` - Disk information tests
- `diskstat/` - Disk statistics tests
- `emit/` - Metrics emission tests
- `gather/` - Data gathering tests
- `metrics/` - Metrics processing tests
- `metrics_computer/` - Metrics computation tests

### Integration Tests
Integration tests are located in the `integration_test/` directory:
- `integration_test.go` - End-to-end tests of the mysql-metrics binary
- Tests real MySQL connectivity and metrics collection
- Validates metrics output format and accuracy

## Troubleshooting

### Docker Issues
- Ensure Docker daemon is running
- Check that your user has permission to access Docker socket

### Go Version Issues
- Ensure Go 1.23.0 or later is installed
- The project uses toolchain 1.24.4, which will be automatically downloaded if needed

### Permission Issues
- Integration tests require access to `/var/run/docker.sock`
- On some systems, you may need to add your user to the `docker` group
- Alternatively, run tests with `sudo` if necessary

### Port Conflicts
- Integration tests use dynamic port allocation
- If you encounter port conflicts, ensure no other MySQL instances are running
- The test setup waits up to 5 minutes for the database to become available

## Dependencies

The testing framework relies on several key dependencies (managed via `go.mod`):
- **github.com/onsi/ginkgo/v2**: BDD testing framework
- **github.com/onsi/gomega**: Matcher library for assertions
- **github.com/DATA-DOG/go-sqlmock**: SQL mocking for unit tests
- **github.com/maxbrunsfeld/counterfeiter/v6**: Interface mocking
- **github.com/go-sql-driver/mysql**: MySQL driver for integration tests

All dependencies are automatically managed by Go modules and do not require manual installation. 
