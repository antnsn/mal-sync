# mal-sync

`mal-sync` is a command-line utility designed to synchronize configurations and rule files for various Grafana Labs observability tools, including Alertmanager, Mimir, and Loki. It leverages the respective command-line tools (`mimirtool`, `lokitool`) to interact with these services, providing a unified interface for managing them, especially in CI/CD pipelines or automated environments.

The tool supports configuration via command-line flags and environment variables, with flags taking precedence.

## Prerequisites

- Docker
- Access to Mimir and/or Loki instances as needed.

## Docker Images

### Official Images

Official Docker images for `mal-sync` are automatically built and published to the GitHub Container Registry (GHCR) by our CI/CD pipeline (see `.github/workflows/docker-publish.yml`).

You can pull the latest stable image (from the `main` branch or version tags) using:

```bash
docker pull ghcr.io/antnsn/mal-sync:latest
# Or for a specific version tag (e.g., v1.0.0):
# docker pull ghcr.io/antnsn/mal-sync:v1.0.0
```

When using these images, replace `mal-sync:dev` in the example `docker run` commands with the appropriate GHCR image tag, e.g., `ghcr.io/antnsn/mal-sync:latest`.

### Local Development Builds

For local development or testing, you can build the Docker image yourself:

```bash
docker build -t mal-sync:dev .
```
This will create a local image tagged `mal-sync:dev`.

## General Usage

```bash
docker run --rm [DOCKER_OPTIONS] mal-sync:dev <subcommand> [SUBCOMMAND_OPTIONS]
```

Replace `[DOCKER_OPTIONS]` with necessary Docker flags (e.g., volume mounts `-v`, environment variables `-e`) and `[SUBCOMMAND_OPTIONS]` with the flags specific to the chosen subcommand.

## Subcommands

`mal-sync` provides the following subcommands:

- `alertmanager`: Syncs Alertmanager configurations.
- `mimir-rules`: Syncs Mimir rule files.
- `loki-rules`: Syncs Loki rule files.

### 1. `alertmanager`

Synchronizes Alertmanager configurations, including the main configuration file and any associated template files, to a Mimir instance (which can act as an Alertmanager).

**Flags & Environment Variables:**

| Flag              | Environment Variable                 | Description                                                                                         | Required | Default     |
| ----------------- | ------------------------------------ | --------------------------------------------------------------------------------------------------- | -------- | ----------- |
| `--config.file`   | `MALSYNC_ALERTMANAGER_CONFIG_FILE`   | Path to the Alertmanager configuration file (e.g., `/config/alertmanager.yaml`).                    | Yes      |             |
| `--templates.dir` | `MALSYNC_ALERTMANAGER_TEMPLATES_DIR` | Path to the directory containing Alertmanager template files (e.g., `/etc/alertmanager/templates`). | No       |             |
| `--mimir.address` | `MALSYNC_ALERTMANAGER_MIMIR_ADDRESS` | Address of the Mimir instance (e.g., `http://mimir-nginx.mimir.svc.cluster.local:80`).              | Yes      |             |
| `--mimir.id`      | `MALSYNC_ALERTMANAGER_MIMIR_ID`      | Mimir tenant ID.                                                                                    | No       | `anonymous` |
| `--temp.dir`      | `MALSYNC_ALERTMANAGER_TEMP_DIR`      | Temporary directory for staging files.                                                              | No       | `/tmp`      |

**Example:**

```bash
docker run --rm \
  -v /path/to/your/alertmanager.yaml:/config/alertmanager.yaml \
  -v /path/to/your/templates:/etc/alertmanager/templates \
  -e MALSYNC_ALERTMANAGER_CONFIG_FILE="/config/alertmanager.yaml" \
  -e MALSYNC_ALERTMANAGER_TEMPLATES_DIR="/etc/alertmanager/templates" \
  -e MALSYNC_ALERTMANAGER_MIMIR_ADDRESS="http://your-mimir-instance:80" \
  -e MALSYNC_ALERTMANAGER_MIMIR_ID="your-tenant-id" \
  mal-sync:dev alertmanager
```

### 2. `mimir-rules`

Synchronizes Mimir rule files to a Mimir instance using `mimirtool rules load`.

**Flags & Environment Variables:**

| Flag                | Environment Variable                 | Description                                                                                | Required | Default     |
| ------------------- | ------------------------------------ | ------------------------------------------------------------------------------------------ | -------- | ----------- |
| `--rules.path`      | `MALSYNC_MIMIRRULES_RULES_PATH`      | Path to a directory containing Mimir rule files (`*.yaml`, `*.yml`) or a single rule file. | Yes      |             |
| `--mimir.address`   | `MALSYNC_MIMIRRULES_MIMIR_ADDRESS`   | Address of the Mimir instance.                                                             | Yes      |             |
| `--mimir.id`        | `MALSYNC_MIMIRRULES_MIMIR_ID`        | Mimir tenant ID.                                                                           | No       | `anonymous` |
| `--rules.namespace` | `MALSYNC_MIMIRRULES_RULES_NAMESPACE` | Mimir namespace to load the rules into.                                                    | Yes      |             |
| `--temp.dir`        | `MALSYNC_MIMIRRULES_TEMP_DIR`        | Temporary directory for staging files.                                                     | No       | `/tmp`      |

**Example:**

```bash
docker run --rm \
  -v /path/to/your/mimir-rules:/rules \
  -e MALSYNC_MIMIRRULES_RULES_PATH="/rules" \
  -e MALSYNC_MIMIRRULES_MIMIR_ADDRESS="http://your-mimir-instance:80" \
  -e MALSYNC_MIMIRRULES_MIMIR_ID="your-tenant-id" \
  -e MALSYNC_MIMIRRULES_RULES_NAMESPACE="your-rules-namespace" \
  mal-sync:dev mimir-rules
```

### 3. `loki-rules`

Synchronizes Loki rule files to a Loki instance using `lokitool rules sync`.

**Flags & Environment Variables:**

| Flag             | Environment Variable             | Description                                                                               | Required | Default |
| ---------------- | -------------------------------- | ----------------------------------------------------------------------------------------- | -------- | ------- |
| `--rules.path`   | `MALSYNC_LOKIRULES_RULES_PATH`   | Path to a directory containing Loki rule files (`*.yaml`, `*.yml`) or a single rule file. | Yes      |         |
| `--loki.address` | `MALSYNC_LOKIRULES_LOKI_ADDRESS` | Address of the Loki instance (e.g., `http://loki.loki.svc.cluster.local:3100`).           | Yes      |         |
| `--loki.org-id`  | `MALSYNC_LOKIRULES_LOKI_ORG_ID`  | Loki Organization ID.                                                                     | Yes      | `fake`  |
| `--temp.dir`     | `MALSYNC_LOKIRULES_TEMP_DIR`     | Temporary directory for staging files.                                                    | No       | `/tmp`  |

**Example:**

```bash
docker run --rm \
  -v /path/to/your/loki-rules:/rules \
  -e MALSYNC_LOKIRULES_RULES_PATH="/rules" \
  -e MALSYNC_LOKIRULES_LOKI_ADDRESS="http://your-loki-instance:3100" \
  -e MALSYNC_LOKIRULES_LOKI_ORG_ID="your-loki-org-id" \
  mal-sync:dev loki-rules
```

## Development

To run linters and tests (TODO: Add tests):

```bash
# go fmt ./...
# go vet ./...
# golangci-lint run
# go test ./...
```

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.
