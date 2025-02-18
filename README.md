# Packwiz Web UI

[![License: MIT](https://img.shields.io/badge/License-MIT-red.svg)](LICENSE)

> [!WARNING]
>
> This is still work-in-progress and is not producing usable builds

A web service to manage [Packwiz](https://github.com/packwiz/packwiz) Minecraft Mod configurations.

You are able to administer Mods by creating new packs and adding, removing or updating mods.
Any changes are immediately available to users.

- Create Modpacks in a beautiful interactive web UI
- Admin user profiles for secure collaboration
- Duplicate existing packs to test out changes
- Serve static Modpack files to users
- Modpack changes are tracked via Git, roll back changes to a previous state
- Upload existing Packwiz mod configurations

## Deploy
This is a web service intended to be deployed as a docker container.
You need to mount a directory into the container to persist your packwiz files.
You can connect the service to an external Postgres database or mount a directory to
persist a local sqlite database.

[Latest Image]()

### Environment Variables

| var             | value                                   | description                                                                                              |
|-----------------|-----------------------------------------|----------------------------------------------------------------------------------------------------------|
| PWW_MODE        | ["production", "development"]           | developers should set this to "development" for additional logging. Do NOT deploy in "development" mode. |
| PWW_PACKWIZ_DIR | absolute path, ie: /packwiz-web/packwiz | this directory contains all the Packwiz toml files and may be configured as a git repository             |
| PWW_DATA_DIR    | absolute path, ie: /packwiz-web/data    | if you are using an sqlite database this is where it will be stored                                      |
|                 |                                         |                                                                                                          |
|                 |                                         |                                                                                                          |

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

---

## Develop

### Requirements
Go (version specified in backend/go.mod)
Node (version specified in frontend/package.json)

```shell
# Build the frontend and backend with:
make build-all

# Run in development mode
make start-dev
```

See readme files for [frontend](frontend/README.md) and [backend](backend/README.md) for specific details about each.

The [examples](examples) directory contains some examples for local deployments.