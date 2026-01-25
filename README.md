# SSH Prometheus Datasource

A Grafana datasource plugin that connects to a Prometheus server through an SSH tunnel.

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
│  ┌──────────────┐       ┌──────────────────────────────────┐                │
│  │              │       │     Grafana Plugin Backend       │                │
│  │   Grafana    │──────▶│                                  │                │
│  │              │ Query │  ┌────────────────────────────┐  │                │
│  └──────────────┘       │  │      SSH Tunnel Client     │  │                │
│                         │  └─────────────┬──────────────┘  │                │
│                         └────────────────┼─────────────────┘                │
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

Coming soon.

### Manual Installation

1. Download the latest release
2. Extract to your Grafana plugins directory
3. Restart Grafana
4. Enable unsigned plugins in `grafana.ini`:
   ```ini
   [plugins]
   allow_loading_unsigned_plugins = tobiasworkstech-ssh-prometheus-datasource
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
docker-compose up
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
docker-compose up
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

## License

Apache 2.0
# tobiasworkstech-ssh-prometheus-datasource
