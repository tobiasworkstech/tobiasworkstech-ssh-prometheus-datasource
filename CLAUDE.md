# SSH Prometheus Datasource Plugin

## Status

**Pending Grafana Plugin Signature** - Submitted and awaiting approval.

## Overview

A Grafana datasource plugin that connects to Prometheus through an SSH tunnel. Useful for accessing Prometheus instances that are not directly accessible from Grafana.

## Architecture

```
Grafana → Plugin Backend (Go) → SSH Tunnel → Remote Prometheus
```

## Project Structure

```
.
├── .github/workflows/     # CI/CD workflows
│   ├── ci.yml            # Build and lint on push/PR
│   └── release.yml       # Build, sign, and release on tags
├── .config/              # Build and lint configuration
│   ├── eslint.config.mjs
│   └── webpack/
├── docker/               # Development environment
│   ├── docker-compose.yml
│   ├── prometheus.yml
│   └── provisioning/
├── src/                  # Frontend React/TypeScript
│   ├── components/
│   ├── datasource/
│   ├── types/
│   ├── img/
│   └── plugin.json
├── pkg/                  # Backend Go code
│   ├── plugin/
│   └── ssh/
├── package.json
├── go.mod
├── CLAUDE.md             # This file
├── README.md             # User documentation
└── CHANGELOG.md          # Version history
```

## Development

### Prerequisites

- Node.js 18+
- Go 1.23+
- Docker & Docker Compose
- Mage (Go build tool): `go install github.com/magefile/mage@latest`

### Building

```bash
# Install dependencies
npm install

# Build frontend
npm run build

# Build backend
go mod tidy
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -o dist/gpx_sshprometheus_datasource_linux_amd64 ./pkg
```

### Development Mode

```bash
# Terminal 1: Frontend watch
npm run dev

# Terminal 2: Start dev environment
cd docker
docker compose up
```

### Testing

1. Open http://localhost:3000 (admin/admin)
2. Add SSH Prometheus datasource
3. Configure:
   - SSH Host: ssh-server
   - SSH Port: 2222
   - SSH Username: testuser
   - Auth Method: Password
   - SSH Password: testpassword
   - Remote Prometheus URL: http://127.0.0.1:9090
4. Save & Test

### Linting

```bash
npm run lint        # Check for issues
npm run lint:fix    # Auto-fix issues
```

ESLint uses flat config format (ESLint 9+) at `.config/eslint.config.mjs`.

## Key Files

| File | Purpose |
|------|---------|
| `src/plugin.json` | Plugin manifest |
| `src/types/index.ts` | TypeScript interfaces |
| `src/components/ConfigEditor.tsx` | SSH config UI |
| `src/components/QueryEditor.tsx` | PromQL editor |
| `src/datasource/datasource.ts` | Frontend datasource |
| `pkg/plugin/datasource.go` | Backend with SSH tunnel + proxy |
| `pkg/ssh/tunnel.go` | SSH connection management |
| `.config/eslint.config.mjs` | ESLint flat config |

## Configuration Options

### SSH Connection
- **SSH Host**: Remote hostname/IP
- **SSH Port**: SSH port (default: 22)
- **SSH Username**: SSH username

### Authentication
- **Password**: Plain password auth
- **SSH Key**: PEM-encoded private key + optional passphrase

### Prometheus
- **Remote Prometheus URL**: URL as seen from remote host (default: http://127.0.0.1:9090)

## CI/CD

### Workflows

- **ci.yml**: Runs on push/PR to main
  - Installs dependencies
  - Lints frontend
  - Runs tests
  - Builds frontend and backend binaries

- **release.yml**: Runs on version tags (v*)
  - Builds plugin
  - Signs plugin (if GRAFANA_ACCESS_POLICY_TOKEN is set)
  - Creates GitHub release with zip artifact
  - Generates build provenance attestation

### Creating a Release

```bash
# Tag a new version
git tag v1.0.1
git push origin v1.0.1
```

The release workflow will automatically build and publish the release.

## Plugin ID

The plugin ID is `tobiasworkstech-sshprometheus-datasource` (must follow pattern `org-name-type`).

---

# Grafana Plugin Development Guide

## Creating a New Grafana Plugin

### Step 1: Scaffold the Plugin

```bash
npx @grafana/create-plugin@latest
```

Choose:
- Plugin type: datasource, panel, or app
- Organization name (used in plugin ID)
- Plugin name
- Backend: yes/no

### Step 2: Plugin ID Format

```
<organization>-<name>-<type>
```

Example: `tobiasworkstech-sshprometheus-datasource`

**Important**: Only 3 segments allowed (org-name-type).

### Step 3: Essential Files

```
├── .config/
│   ├── eslint.config.mjs    # ESLint 9 flat config
│   └── webpack/
├── src/
│   ├── plugin.json          # Manifest (required)
│   ├── module.ts            # Entry point
│   ├── components/          # React components
│   ├── datasource/          # Datasource logic
│   ├── types/               # TypeScript types
│   └── img/
│       └── logo.svg         # Plugin logo
├── pkg/                     # Go backend (if needed)
├── package.json
└── go.mod                   # If backend
```

### Step 4: Build Commands

```bash
# Frontend
npm install
npm run build
npm run dev        # watch mode

# Backend
go mod tidy
go build -o dist/gpx_name_linux_amd64 ./pkg
```

## Plugin Signing Process

### Signature Levels

| Level | Use Case |
|-------|----------|
| Private | Internal org use only |
| Community | Open-source public plugins |
| Commercial | Paid/proprietary plugins |

### Requirements for Community Signing

1. **Public GitHub repository**
2. **Apache 2.0 license** (or compatible)
3. **Complete plugin.json** with all metadata
4. **README.md** with documentation
5. **Backend binaries** for linux/amd64, darwin/amd64, windows/amd64
6. **Screenshots** in plugin.json for catalog display

### Submission Process

1. **Create Grafana.com account**: https://grafana.com/auth/sign-up
2. **Create organization**: https://grafana.com/orgs/new
3. **Validate plugin**:
   ```bash
   go install github.com/grafana/plugin-validator/cmd/plugincheck@latest
   plugincheck -sourceCodeUri file://. dist/
   ```
4. **Create GitHub release** with zip containing built plugin
5. **Submit via Grafana.com**: https://grafana.com/orgs/YOUR_ORG/plugins

### plugin.json Template

```json
{
  "$schema": "https://raw.githubusercontent.com/grafana/grafana/main/docs/sources/developers/plugins/plugin.schema.json",
  "type": "datasource",
  "name": "Plugin Display Name",
  "id": "org-name-datasource",
  "backend": true,
  "executable": "gpx_name",
  "alerting": true,
  "metrics": true,
  "info": {
    "description": "What the plugin does",
    "author": {
      "name": "Author Name",
      "url": "https://example.com"
    },
    "keywords": ["keyword1", "keyword2"],
    "logos": {
      "small": "img/logo.svg",
      "large": "img/logo.svg"
    },
    "screenshots": [
      {
        "name": "Screenshot Name",
        "path": "img/screenshot.png"
      }
    ],
    "version": "1.0.0",
    "updated": "2024-01-01"
  },
  "dependencies": {
    "grafanaDependency": ">=9.0.0",
    "plugins": []
  }
}
```

### Release Zip Structure

```
pluginid/
├── plugin.json
├── module.js
├── gpx_name_linux_amd64
├── gpx_name_darwin_amd64
├── gpx_name_windows_amd64.exe
├── img/
│   ├── logo.svg
│   └── screenshot.png
└── README.md
```

### Common Validation Errors

| Error | Solution |
|-------|----------|
| `invalid-metadata` | Plugin ID must match pattern `org-name-type` (3 segments) |
| `go-mod-not-found` | Place go.mod at repository root, not in subdirectory |
| `packagejson-not-found` | Place package.json at repository root |
| `code-rules-access-os-environment` | Don't include Magefile.go in source (uses os.Environ) |
| `js-map-no-match` | Source code must match built assets exactly |

### After Approval

- Plugin appears in Grafana catalog
- Install: `grafana-cli plugins install plugin-id`
- No `allow_loading_unsigned_plugins` needed

### Publishing Updates

1. Increment version in `plugin.json`
2. Create new GitHub release
3. Submit new version via Grafana.com

## Resources

- [Plugin Tools](https://grafana.com/developers/plugin-tools/)
- [Plugin Validator](https://github.com/grafana/plugin-validator)
- [Publish Guide](https://grafana.com/developers/plugin-tools/publish-a-plugin/publish-a-plugin)
- [Community Forums](https://community.grafana.com/)
