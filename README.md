# SSH Prometheus Datasource

A Grafana datasource plugin that connects to a Prometheus server through an SSH tunnel.

## Status

**Pending Grafana Plugin Signature** - Submitted and awaiting approval. Once approved, install via `grafana-cli plugins install tobiasworkstech-sshprometheus-datasource`.

## Use Case

This plugin is useful when:
- Your Prometheus instance is behind a firewall
- Prometheus is only accessible from specific jump hosts
- You need to access Prometheus in a private network

## How It Works

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              YOUR NETWORK                                   │
│                                                                             │
│  ┌──────────────┐       ┌──────────────────────────────────────┐            │
│  │              │       │     Grafana Plugin Backend           │            │
│  │   Grafana    │──────▶│                                      │            │
│  │              │ Query │  ┌────────────────────────────────┐  │            │
│  └──────────────┘       │  │      SSH Tunnel Client         │  │            │
│                         │  └─────────────┬──────────────────┘  │            │
│                         └────────────────┼───────────────────────┘          │
│                                          │                                  │
└──────────────────────────────────────────┼──────────────────────────────────┘
                                           │ SSH Connection (Port 22)
                                           │
┌──────────────────────────────────────────┼──────────────────────────────────┐
│                            PRIVATE NETWORK / FIREWALL                       │
│                                          │                                  │
│                         ┌────────────────▼─────────────────┐                │
│                         │                                  │                │
│                         │        SSH Server / Jump Host    │                │
│                         │                                  │                │
│                         └────────────────┬─────────────────┘                │
│                                          │ Local Connection                 │
│                                          │ (http://127.0.0.1:9090)          │
│                         ┌────────────────▼─────────────────┐                │
│                         │                                  │                │
│                         │          Prometheus              │                │
│                         │                                  │                │
│                         └──────────────────────────────────┘                │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Flow:**
1. Grafana sends PromQL queries to the plugin backend
2. The plugin establishes an SSH tunnel to the remote SSH server
3. Queries are forwarded through the tunnel to Prometheus
4. Results return through the same secure tunnel

## Features

- SSH tunnel to remote Prometheus servers
- Support for password and SSH key authentication
- Full Prometheus query support (range queries, instant queries)
- Variable support with `label_values()`, `label_names()`, `metrics()`
- Alerting support

## Installation

### From Grafana Marketplace

Coming soon (pending plugin signature approval).

### Manual Installation

1. Download the latest release
2. Extract to your Grafana plugins directory
3. Restart Grafana
4. Enable unsigned plugins in `grafana.ini`:
   ```ini
   [plugins]
   allow_loading_unsigned_plugins = tobiasworkstech-sshprometheus-datasource
   ```

## Configuration

### SSH Connection

| Field | Description |
|-------|-------------|
| SSH Host | Hostname or IP of the SSH server |
| SSH Port | SSH port (default: 22) |
| SSH Username | SSH username for authentication |

### Authentication

Choose between:

**Password Authentication**
- SSH Password: Your SSH password

**Key-Based Authentication**
- SSH Private Key: PEM-encoded private key (contents of `~/.ssh/id_rsa`)
- Key Passphrase: Optional passphrase if key is encrypted

### Prometheus Settings

| Field | Description |
|-------|-------------|
| Remote Prometheus URL | URL of Prometheus as seen from the SSH host (default: http://127.0.0.1:9090) |

## Query Editor

The query editor supports standard PromQL:

- **Expression**: PromQL query (e.g., `up`, `rate(http_requests_total[5m])`)
- **Legend**: Format using `{{label}}` syntax
- **Min Interval**: Minimum step interval
- **Instant**: Toggle for instant vs range queries

## Variable Support

Use these functions in variable queries:

- `label_values(label_name)` - Get all values for a label
- `label_values(metric, label_name)` - Get label values for a specific metric
- `label_names()` - Get all label names
- `metrics(filter)` - Get metric names matching regex

## Development

### Prerequisites

- Node.js 18+
- Go 1.23+
- Docker & Docker Compose
- Mage (Go build tool): `go install github.com/magefile/mage@latest`

### Running the Dev Environment

The project includes a Docker Compose setup with Grafana, Prometheus, and an SSH server for testing.

```bash
# Build the plugin first
cd plugin
npm install
npm run build
go mod tidy
mage -v build:linux

# Start the dev environment
cd ../docker
docker compose up
```

Once running, access:
- **Grafana**: http://localhost:3000 (login: admin/admin)
- **Prometheus**: http://localhost:9090

### Test Datasource Configuration

The dev environment includes a pre-configured SSH server:

| Setting | Value |
|---------|-------|
| SSH Host | ssh-server |
| SSH Port | 2222 |
| SSH Username | testuser |
| SSH Password | testpassword |
| Remote Prometheus URL | http://prometheus:9090 |

### Frontend Watch Mode

For live frontend development:

```bash
# Terminal 1: Watch frontend changes
cd plugin
npm run dev

# Terminal 2: Run docker environment
cd docker
docker compose up
```

## Building

```bash
# Install dependencies
cd plugin
npm install
go mod tidy

# Build frontend
npm run build

# Build backend (requires mage)
mage -v build:linux
mage -v build:darwin
mage -v build:windows
```

---

# Guide: Creating and Signing a Grafana Plugin

This section documents the complete process for creating a new Grafana plugin and getting it officially signed.

## Overview

Grafana requires plugins to be signed before they can be loaded without special configuration. There are three signature levels:
- **Private**: For internal use within an organization
- **Community**: For open-source plugins shared publicly
- **Commercial**: For paid/proprietary plugins

## Step 1: Create the Plugin

### Using the Grafana Plugin Tools

```bash
# Install the create-plugin tool
npx @grafana/create-plugin@latest

# Follow the prompts:
# - Choose plugin type (datasource, panel, app)
# - Enter plugin name
# - Enter organization name
# - Choose if backend is needed
```

### Plugin ID Convention

Your plugin ID must follow this format:
```
<organization>-<plugin-name>-<plugin-type>
```
Example: `tobiasworkstech-sshprometheus-datasource`

### Required Files

Ensure your plugin has:
- `src/plugin.json` - Plugin manifest with correct metadata
- `LICENSE` - Apache 2.0 or compatible license (required for community signing)
- `README.md` - Documentation for users
- Logo images in `src/img/`

### plugin.json Requirements

```json
{
  "$schema": "https://raw.githubusercontent.com/grafana/grafana/main/docs/sources/developers/plugins/plugin.schema.json",
  "type": "datasource",
  "name": "Your Plugin Name",
  "id": "yourorg-pluginname-datasource",
  "backend": true,
  "executable": "gpx_pluginname",
  "info": {
    "description": "Clear description of what the plugin does",
    "author": {
      "name": "Your Name or Organization",
      "url": "https://yourwebsite.com"
    },
    "keywords": ["relevant", "keywords"],
    "logos": {
      "small": "img/logo.svg",
      "large": "img/logo.svg"
    },
    "version": "1.0.0",
    "updated": "2025-01-27"
  },
  "dependencies": {
    "grafanaDependency": ">=9.0.0",
    "plugins": []
  }
}
```

## Step 2: Build and Validate

### Build the Plugin

```bash
cd plugin

# Frontend
npm install
npm run build

# Backend (if applicable)
go mod tidy
mage -v build:linux
mage -v build:darwin
mage -v build:windows
```

### Validate with plugincheck

```bash
# Install plugincheck
go install github.com/grafana/plugin-validator/cmd/plugincheck@latest

# Run validation
plugincheck -sourceCodeUri file://. dist/
```

Fix any errors before proceeding.

## Step 3: Create a GitHub Repository

1. Create a public GitHub repository (required for community plugins)
2. Push your plugin code
3. Create a release with the built plugin as a zip archive

### Release Zip Structure

Your release zip should contain:
```
yourorg-pluginname-datasource/
├── plugin.json
├── module.js
├── module.js.map (optional)
├── gpx_pluginname_linux_amd64 (backend binary)
├── gpx_pluginname_darwin_amd64 (backend binary)
├── gpx_pluginname_windows_amd64.exe (backend binary)
├── img/
│   └── logo.svg
└── README.md
```

## Step 4: Register on Grafana.com

1. Go to https://grafana.com/auth/sign-up and create an account
2. Navigate to https://grafana.com/orgs/new to create an organization
3. Go to **My Account** > **API Keys** and generate a Cloud Access Policy token with plugin publishing permissions

## Step 5: Submit for Signing

### Using the Grafana Plugin CLI

```bash
# Install the grafana-toolkit (if not using newer tools)
npm install -g @grafana/toolkit

# Or use npx with the signing tool directly
npx @grafana/sign-plugin@latest
```

### Submit via Grafana.com

1. Go to https://grafana.com/orgs/YOUR_ORG/plugins
2. Click "Submit Plugin"
3. Provide:
   - GitHub repository URL
   - Plugin ID (must match plugin.json)
   - Download URL for the release zip
   - Plugin type
   - Signature level (usually "community" for open-source)

### Submission Requirements Checklist

- [ ] Plugin ID follows naming convention
- [ ] `plugin.json` has all required fields
- [ ] README.md documents usage
- [ ] LICENSE file present (Apache 2.0 or compatible)
- [ ] No security vulnerabilities
- [ ] Backend plugins include binaries for linux/amd64, darwin/amd64, windows/amd64
- [ ] Plugin passes `plugincheck` validation
- [ ] GitHub repository is public
- [ ] Release zip is publicly downloadable

## Step 6: Review Process

After submission:

1. **Automated checks** run immediately (takes minutes)
2. **Manual review** by Grafana team (takes days to weeks)
3. You may receive feedback requiring changes
4. Once approved, your plugin is signed and published

### Common Review Feedback

- Missing or incomplete documentation
- Security concerns in code
- Incorrect plugin.json metadata
- Missing backend binaries
- License issues

## Step 7: After Approval

Once signed:

1. Your plugin appears in the Grafana plugin catalog
2. Users can install via: `grafana-cli plugins install yourorg-pluginname-datasource`
3. No more `allow_loading_unsigned_plugins` needed

### Publishing Updates

For updates:
1. Increment version in `plugin.json`
2. Create new GitHub release
3. Submit new version through Grafana.com
4. New versions go through review (usually faster)

## Tips for Successful Signing

1. **Test thoroughly** - Use the plugin in multiple Grafana versions
2. **Clear documentation** - Explain all configuration options
3. **Handle errors gracefully** - Don't crash on bad input
4. **Follow security best practices** - Never log sensitive data
5. **Responsive to feedback** - Address review comments promptly

## Resources

- [Grafana Plugin Documentation](https://grafana.com/developers/plugin-tools/)
- [Plugin Submission Guide](https://grafana.com/developers/plugin-tools/publish-a-plugin/publish-a-plugin)
- [Plugin Validator](https://github.com/grafana/plugin-validator)
- [Grafana Community Forums](https://community.grafana.com/)

## License

Apache 2.0
