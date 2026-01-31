# Claude Skills for SSH Prometheus Datasource

## Quick Commands

### Build

```bash
# Install dependencies
npm install

# Frontend only
npm run build

# Backend (all platforms)
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -o dist/gpx_sshprometheus_datasource_linux_amd64 ./pkg
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-w -s" -o dist/gpx_sshprometheus_datasource_linux_arm64 ./pkg
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -o dist/gpx_sshprometheus_datasource_darwin_amd64 ./pkg
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-w -s" -o dist/gpx_sshprometheus_datasource_darwin_arm64 ./pkg
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -o dist/gpx_sshprometheus_datasource_windows_amd64.exe ./pkg
```

### Lint

```bash
npm run lint
npm run lint:fix
```

### Test

```bash
# Frontend tests
npm run test

# Backend tests
go test -v ./...
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
curl -s -u admin:admin http://localhost:3000/api/plugins/tobiasworkstech-sshprometheus-datasource/settings | jq .

# Create datasource
curl -s -X POST -u admin:admin http://localhost:3000/api/datasources \
  -H "Content-Type: application/json" \
  -d '{"name":"SSH Prometheus Test","type":"tobiasworkstech-sshprometheus-datasource","access":"proxy","jsonData":{"sshHost":"ssh-server","sshPort":2222,"sshUsername":"testuser","authMethod":"password","prometheusUrl":"http://127.0.0.1:9090"},"secureJsonData":{"sshPassword":"testpassword"}}'

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
.
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

### Plugin ID
`tobiasworkstech-sshprometheus-datasource`

### Dev Environment Credentials
- Grafana: http://localhost:3000 (admin/admin)
- SSH: ssh-server:2222 (testuser/testpassword)
- Prometheus: http://127.0.0.1:9090 (via SSH tunnel)

### ESLint
Uses flat config format at `.config/eslint.config.mjs`. The lint script includes `--config` flag.

### CI/CD
- `ci.yml`: Runs on push/PR to main
- `release.yml`: Runs on version tags, creates GitHub release
