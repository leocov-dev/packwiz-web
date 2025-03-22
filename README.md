# Packwiz Web UI

[![License: MIT](https://img.shields.io/badge/License-MIT-red.svg)](LICENSE)

> [!WARNING]
>
> **This is still work-in-progress**

A web service to manage [Packwiz](https://github.com/packwiz/packwiz) Minecraft Mod configurations.

You are able to administer Mods by creating new packs and adding, removing or updating mods.
Any changes are immediately available to users.

1. [x] Create Modpacks in a beautiful interactive web UI
2. [x] Admin and User accounts for secure collaboration
3. [x] Serve static Modpack files to users
4. [ ] Duplicate existing packs to test out changes
5. [ ] Modpack changes are tracked via Git, roll back changes to a previous state
6. [ ] Upload existing Packwiz mod configurations

## Deploy
This is a web service intended to be deployed as a docker container.
You need to mount a directory into the container to persist your packwiz files.
You can connect the service to an external Postgres database or a local sqlite database.

[Latest Container Image]()

### Environment Variables

| var                 | value                                   | description                                                                                              |
|---------------------|-----------------------------------------|----------------------------------------------------------------------------------------------------------|
| PWW_MODE            | ["production", "development"]           | developers should set this to "development" for additional logging. Do NOT deploy in "development" mode. |
| PWW_PACKWIZ_DIR     | absolute path, ie: /packwiz-web/packwiz | this directory contains all the Packwiz toml files and may be configured as a git repository             |
| PWW_DATA_DIR        | absolute path, ie: /packwiz-web/data    | if you are using an sqlite database this is where it will be stored                                      |
| PWW_ADMIN_PASSWORD  | min 16 char string                      | set the password for the default `admin` account                                                         |
| PWW_DATABASE        | ["postres", "sqlite"]                   | set the database to use on the backend                                                                   |
| PWW_SESSION_SECRET  | a long random string                    | encryption key for the HTTP session                                                                      |
| PWW_TRUSTED_PROXIES | comma separated string list             | `gin` server trusted proxies configuration, set to your public host if behind a reverse proxy            |

### User Access

Users access mods at the static file endpoint:
`https://<host>/packwiz/<pack-name>/pack.toml`

If a modpack is nod public each user must use their access token:
`https://<host>/packwiz/<pack-name>/pack.toml?token=<access-token>`
Each access is logged and available to admins in the UI.


### Security

By default, admins may create user accounts with passwords managed by the service.
Additionally, admins may configure OIDC providers and a user can choose to link their account to one.

The static file server is reasonably secure against directory traversal attacks and can
only serve files directly inside the packwiz dir.

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
