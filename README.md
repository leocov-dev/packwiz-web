# Packwiz Web UI

[![License: MIT](https://img.shields.io/badge/License-MIT-red.svg)](LICENSE)

> [!WARNING]
>
> **This is still a work-in-progress**
> Some features may be broken or might change drastically between releases

A web service to manage [Packwiz](https://github.com/packwiz/packwiz) Minecraft Mod configurations.
This uses a fork of Packwiz, [packwiz-nxt](https://github.com/leocov-dev/packwiz-nxt) that exposes more functionality as a library.

You are able to administer Mods by creating new packs and adding, removing or updating mods.
Any changes are immediately available to users.

1. [ ] Manage Modpacks in a beautiful interactive web UI
   1. [x] Create Packs and add Mods
   2. [ ] Edit Packs and Mods
   3. [ ] Update Packs and Mods
2. [x] Admin and User accounts for secure collaboration
3. [x] Serve static Modpack files to servers/clients
4. [ ] Duplicate existing packs to test out changes
5. [ ] Snapshot Modpacks and roll back to previous states
6. [ ] Import existing Packwiz mod configurations
7. [ ] OIDC authentication

## Deploy
This is a web service intended to be deployed as a docker container.
A Postgres database is required. 
See [docker-compose.yml](examples/docker-compose/docker-compose.yml) for a very basic example.

[Latest Container Image](https://github.com/leocov-dev/packwiz-web/pkgs/container/packwiz-web)

### Environment Variables

| var                 | value                                  | description                                                                                                                           |
|---------------------|----------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------|
| PWW_MODE            | ["production", "development"]          | Developers should set this to `development` for additional logging. Do NOT deploy in `development` mode. The default is `production`. |
| PWW_ADMIN_PASSWORD  | min 16 char string                     | Set the password for the default `admin` account, when starting the container this will always be applied to the admin account.       |
| PWW_SESSION_SECRET  | a long random string                   | Encryption key for the HTTP session. You must set this, there is no default.                                                          |
| PWW_TRUSTED_PROXIES | comma separated string list            | The `gin` server trusted proxies configuration, set to your public host if behind a reverse proxy.                                    |
| PWW_CF_API_KEY      | base64 encoded Curseforge API key      | In order to register curseforge mods you must have an API key. The pre-build container images already include one by default.         |
| PWW_GH_API_KEY      | GitHub API key                         | To avoid rate limits or download from private repositories from GitHub you can supply an API key. None is included by default.        |

Postgres connection vars:

| var             | description                                                  |
|-----------------|--------------------------------------------------------------|
| PWW_PG_HOST     | database host, url, ip addr, default: localhost              |
| PWW_PG_PORT     | database connection port, default: 5432                      |
| PWW_PG_USER     | database connection username, default: postgres              |
| PWW_PG_PASSWORD | database connection password                                 |
| PWW_PG_DBNAME   | database name to use, should already exist, default: packwiz |


### User Access

Users access mods at the static file endpoint:
- if the pack is public:
  - `https://<host>/packwiz/public/<pack-name>/pack.toml`
  - this url may be shared with anyone
- if the pack is not public:
  - `https://<host>/packwiz/<user-token>/<pack-name>/pack.toml`
  - this url is not intended to be shared
  - user access is logged 
  - the token can be regenerated

### Security

By default, admins may create user accounts with passwords managed by the service.

#### Audit Logs

All API actions are logged in an audit log table.

---

## Develop

### Requirements
 - Go (version specified in backend/go.mod)
 - Node (version specified in frontend/package.json)

```shell
# Build the frontend and backend with:
make build-all

# set or export the minimum env vars
PWW_MODE="development"
PWW_PG_PASSWORD="yourdbpass"

# Run in development mode
make start-dev
```

See readme files for [frontend](frontend/README.md) and [backend](backend/README.md) for specific details about each.

The [examples](examples) directory contains some examples for local deployments.

### Container
The frontend and backend can be built into a container image for deployment.

```shell
# build and run the container locally
make build-image
docker run packwiz-web
```
