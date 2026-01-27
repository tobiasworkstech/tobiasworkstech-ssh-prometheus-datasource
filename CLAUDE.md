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

- `plugin/src/` - Frontend React/TypeScript code
- `plugin/pkg/` - Backend Go code
- `docker/` - Development environment

## Development

### Prerequisites

- Node.js 18+
- Go 1.21+
- Docker & Docker Compose
- Mage (Go build tool)

### Building

```bash
# Frontend
cd plugin
npm install
npm run build

# Backend
cd plugin
go mod tidy
mage -v build:linux  # or build:darwin, build:windows
```

### Development Mode

```bash
# Terminal 1: Frontend watch
cd plugin
npm run dev

# Terminal 2: Start dev environment
cd docker
docker-compose up
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
   - Remote Prometheus URL: http://prometheus:9090
4. Save & Test

## Key Files

| File | Purpose |
|------|---------|
| `plugin/src/plugin.json` | Plugin manifest |
| `plugin/src/types/index.ts` | TypeScript interfaces |
| `plugin/src/components/ConfigEditor.tsx` | SSH config UI |
| `plugin/src/components/QueryEditor.tsx` | PromQL editor |
| `plugin/src/datasource/datasource.ts` | Frontend datasource |
| `plugin/pkg/plugin/datasource.go` | Backend with SSH tunnel + proxy |
| `plugin/pkg/ssh/tunnel.go` | SSH connection management |

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
<organization>-<plugin-name>-<plugin-type>
```

Example: `tobiasworkstech-ssh-prometheus-datasource`

### Step 3: Essential Files

```
plugin/
├── src/
│   ├── plugin.json          # Manifest (required)
│   ├── module.ts             # Entry point
│   ├── components/           # React components
│   ├── datasource/           # Datasource logic
│   ├── types/                # TypeScript types
│   └── img/
│       └── logo.svg          # Plugin logo
├── pkg/                      # Go backend (if needed)
├── package.json
├── go.mod                    # If backend
└── Magefile.go              # If backend
```

### Step 4: Build Commands

```bash
# Frontend
npm install
npm run build
npm run dev        # watch mode

# Backend
go mod tidy
mage -v build:linux
mage -v build:darwin
mage -v build:windows
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
  "id": "org-name-type",
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
│   └── logo.svg
└── README.md
```

### Review Timeline

- Automated checks: Minutes
- Manual review: Days to weeks
- Updates: Usually faster review

### After Approval

- Plugin appears in Grafana catalog
- Install: `grafana-cli plugins install plugin-id`
- No `allow_loading_unsigned_plugins` needed

### Publishing Updates

1. Increment version in `plugin.json`
2. Create new GitHub release
3. Submit new version via Grafana.com

## Common Issues

| Issue | Solution |
|-------|----------|
| Plugin not loading | Check `allow_loading_unsigned_plugins` in grafana.ini |
| Backend not starting | Check binary permissions and logs |
| Validation errors | Run `plugincheck` and fix reported issues |
| Signing rejected | Review feedback, fix issues, resubmit |

## Resources

- [Plugin Tools](https://grafana.com/developers/plugin-tools/)
- [Plugin Validator](https://github.com/grafana/plugin-validator)
- [Publish Guide](https://grafana.com/docs/grafana/latest/developers/plugins/publish-a-plugin/)
- [Community Forums](https://community.grafana.com/)
