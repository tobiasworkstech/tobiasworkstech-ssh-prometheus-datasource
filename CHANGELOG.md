# Changelog

All notable changes to the Grafana SSH Prometheus Datasource plugin will be documented in this file.

## [1.0.1] - 2026-01-27

### Changed

- **BREAKING**: Renamed plugin ID from `tobiasworkstech-ssh-prometheus-datasource` to `tobiasworkstech-sshprometheus-datasource` to comply with Grafana plugin ID format (org-name-type pattern)
- Restructured repository: moved plugin files from `plugin/` subdirectory to root level
- Updated executable name to `gpx_sshprometheus_datasource`
- Removed Magefile.go (contained disallowed operations for Grafana plugin validation)

### Fixed

- Fixed plugin.json ID pattern validation error
- Fixed go.mod/package.json not found errors (files now at root)
- Fixed broken documentation link

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
