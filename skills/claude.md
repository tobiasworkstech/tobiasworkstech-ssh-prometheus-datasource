# Claude Skills for SSH Prometheus Datasource

## Quick Commands

### Build

```bash
# Frontend only
cd plugin && npm run build

# Backend only (current OS)
cd plugin && go build -o dist/gpx_ssh_prometheus_datasource ./pkg

# All platforms
cd plugin && mage -v build:linux && mage -v build:darwin && mage -v build:windows
```

### Lint

```bash
cd plugin && npm run lint
cd plugin && npm run lint:fix
```

### Test

```bash
# Frontend tests
cd plugin && npm run test

# Backend tests
cd plugin && go test -v ./...
```

### Development Environment

```bash
# Start
cd docker && docker compose up -d

# Stop
cd docker && docker compose down

# View logs
docker logs docker-grafana-1 -f
```

### Test Datasource via API

```bash
# Check plugin is loaded
curl -s -u admin:admin http://localhost:3000/api/plugins/tobiasworkstech-ssh-prometheus-datasource/settings | jq .

# Create datasource
curl -s -X POST -u admin:admin http://localhost:3000/api/datasources \
  -H "Content-Type: application/json" \
  -d '{"name":"SSH Prometheus Test","type":"tobiasworkstech-ssh-prometheus-datasource","access":"proxy","jsonData":{"sshHost":"ssh-server","sshPort":2222,"sshUsername":"testuser","authMethod":"password","prometheusUrl":"http://127.0.0.1:9090"},"secureJsonData":{"sshPassword":"testpassword"}}'

# Test datasource health (replace UID)
curl -s -u admin:admin http://localhost:3000/api/datasources/uid/<UID>/health | jq .
```

### Release

```bash
# Tag and push (triggers release workflow)
git tag v1.0.1
git push origin v1.0.1
```

## Project Structure

```
plugin/
├── .config/eslint.config.mjs  # ESLint flat config (ESLint 9+)
├── src/                       # Frontend (React/TypeScript)
│   ├── components/           # UI components
│   ├── datasource/           # Datasource logic
│   └── plugin.json           # Plugin manifest
├── pkg/                       # Backend (Go)
│   ├── plugin/               # Grafana plugin interface
│   └── ssh/                  # SSH tunnel management
└── dist/                      # Build output
```

## Key Configuration

### Dev Environment Credentials
- Grafana: http://localhost:3000 (admin/admin)
- SSH: ssh-server:2222 (testuser/testpassword)
- Prometheus: http://127.0.0.1:9090 (via SSH tunnel)

### ESLint
Uses flat config format at `plugin/.config/eslint.config.mjs`. The lint script includes `--config` flag.

### CI/CD
- `ci.yml`: Runs on push/PR to main
- `release.yml`: Runs on version tags, creates GitHub release
