# Deployment Examples

This directory provides Docker Compose templates for deploying `packwiz-web`.

## Choosing an Architecture

| Template | Job Processing | Resource Usage | Operational Complexity | Recommended Use Case |
| :--- | :--- | :--- | :--- | :--- |
| **[`docker-compose/`](docker-compose/docker-compose.yml)** | In-process (`--worker`) | Low (single application container) | Minimal | Small deployments, personal instances, homelabs |
| **[`docker-compose-worker/`](docker-compose-worker/docker-compose.yml)** | Dedicated container (`worker`) | Decoupled (isolated web & worker containers) | Moderate | Production, multi-user teams, heavy modpack migration workloads |

---

### 1. Basic / All-in-One (`docker-compose/`)
The web service runs both the HTTP server and the background job processor in the same container process by specifying `command: ["start", "--migrate", "--worker"]`.

- **Pros**: Minimal footprint, lowest memory overhead, single container to monitor.
- **Cons**: CPU-intensive background tasks (such as checking mod updates or running modpack migrations) share resources with HTTP request handling.

```shell
cd examples/docker-compose
docker compose up -d
```

---

### 2. Dedicated Worker (`docker-compose-worker/`)
The web container runs `command: ["start", "--migrate"]` (without `--worker`), which disables in-process background job processing on the web node and reserves it exclusively for handling web and API requests. A separate container (`packwiz-worker`) runs `command: ["worker"]` using the same image to poll and process River queue jobs from PostgreSQL.

- **Pros**: Fault and resource isolation; heavy background processing will not starve the web interface of CPU or memory; web and worker tiers can be scaled independently.
- **Cons**: Requires additional container management and baseline memory overhead for two container runtimes.

```shell
cd examples/docker-compose-worker
docker compose up -d
```
