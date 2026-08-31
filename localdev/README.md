# Local Development

This directory holds a Postgres-only [docker-compose.yml](docker-compose.yml)
for local development — it is **not** a deployment example (see
[examples](../examples) for that).

Running `make start-dev` from the repo root starts this database
automatically and points the backend at it, so in most cases you don't need
to touch this directory directly.

To manage the database manually:

```shell
# start
docker compose -f localdev/docker-compose.yml up -d

# stop
docker compose -f localdev/docker-compose.yml down
```

The database is exposed on `localhost:55432` (a non-default port, so it
won't collide with any Postgres you already have running locally).
