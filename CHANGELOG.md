# Changelog

All notable changes to the Grafana SSH Prometheus Datasource plugin will be documented in this file.

## [1.0.0] - 2025-01-27

### Added

- Initial release of the SSH Prometheus Datasource plugin
- Connect to Prometheus through an SSH tunnel for secure access
- Support for both password and SSH key authentication
- Optional passphrase support for encrypted SSH keys
- Full PromQL query support with Grafana's native query editor
- Alerting support through Grafana's unified alerting
- Template variable support for dynamic dashboards
- Configurable options:
  - SSH host and port configuration
  - SSH username and authentication method
  - Remote Prometheus URL (as seen from the SSH host)
- Multi-platform backend support (Linux, macOS, Windows on amd64/arm64)
- Docker development environment for easy testing
