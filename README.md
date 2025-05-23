# Packwiz Web UI

[![License: MIT](https://img.shields.io/badge/License-MIT-red.svg)](LICENSE)

> [!WARNING]
>
> **This is still work-in-progress, but will run locally**

A web service to manage [Packwiz](https://github.com/packwiz/packwiz) Minecraft Mod configurations.

You are able to administer Mods by creating new packs and adding, removing or updating mods.
Any changes are immediately available to users.

1. [x] Create Modpacks in a beautiful interactive web UI
2. [x] Admin and User accounts for secure collaboration
3. [x] Serve static Modpack files to users
4. [ ] Duplicate existing packs to test out changes
5. [ ] Modpack changes are tracked via Git. Roll back changes to a previous state
6. [ ] Upload existing Packwiz mod configurations

## Deploy
This is a web service intended to be deployed as a docker container.
You need to mount a directory into the container to persist your packwiz files.
You can connect the service to an external Postgres database or a sqlite database in a mounted directory.

[Latest Container Image]()

### Environment Variables

| var                 | value                                  | description                                                                                                                          |
|---------------------|----------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------|
| PWW_MODE            | ["production", "development"]          | Developers should set this to `development` for additional logging. Do NOT deploy in `development` mode. The default is `production` |
| PWW_ADMIN_PASSWORD  | min 16 char string                     | Set the password for the default `admin` account, when starting the container this will always be applied to the admin account       |
| PWW_DATABASE        | ["postres", "sqlite"]                  | Set the database to use on the backend, the default is `sqlite`                                                                      |
| PWW_DATA_DIR        | absolute path, ie: `/packwiz-web/data` | If you are using an sqlite database this is where it will be stored                                                                  |
| PWW_SESSION_SECRET  | a long random string                   | Encryption key for the HTTP session. You must set this, there is no default.                                                         |
| PWW_TRUSTED_PROXIES | comma separated string list            | The `gin` server trusted proxies configuration, set to your public host if behind a reverse proxy.                                   |
| PWW_CF_API_KEY      | base64 encoded Curseforge API key      | In order to register curseforge mods you must have an API key. The pre-build container images already include one by default.        |
| PWW_GH_API_KEY      | GitHub API key                         | To avoid rate limits if registering many mods from GitHub you can supply an API key. None is included by default.                    |

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

# Run in development mode
make start-dev
```

See readme files for [frontend](frontend/README.md) and [backend](backend/README.md) for specific details about each.

The [examples](examples) directory contains some examples for local deployments.

### Container
The frontend and backend can be built into a container image for deployment.

```shell
# build an image locally
make build-image

# basic run (env vars and volumes must be set)
docker run --rm packwiz-web start
```
