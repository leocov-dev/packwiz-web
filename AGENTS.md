# AGENTS.md

Packwiz Web UI: a web service to manage [Packwiz](https://github.com/packwiz/packwiz) Minecraft mod packs.
Create packs, add and update mods, and serve the result to servers and clients.

## Layout

| Path        | Description                                                  |
|-------------|--------------------------------------------------------------|
| `backend/`  | Go service (API, jobs, static pack file serving, Postgres)   |
| `frontend/` | Node web UI                                                  |
| `docs/`     | User-facing documentation (OIDC, reverse proxy, screenshots) |
| `examples/` | Example docker-compose deployments                           |
| `localdev/` | Local development Postgres via Docker                        |
| `Makefile`  | Build and dev entry points                                   |
| `Dockerfile`| Builds frontend and backend into a single container image    |

## Instructions

Detailed instructions live in the respective directories. Read the relevant file before working in that area:

- Backend (Go): [backend/AGENTS.md](backend/AGENTS.md)
- Frontend: [frontend/AGENTS.md](frontend/AGENTS.md)

## Common Commands

```shell
make build-all   # build frontend and backend
make start-dev   # run in development mode (requires Docker for local Postgres)
make dev-db-down # stop the local dev database
```

See [README.md](README.md) for deployment, configuration and general project details.
