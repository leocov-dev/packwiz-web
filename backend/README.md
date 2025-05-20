# packwiz-web Backend

## Dependencies

- Go as specified in [go.mod](go.mod)


## Local

```
# see available commands
make

# build the backend
make build

# run dev server locally
make start-dev

# run unit tests
make test
```


## Developer

- [Server Router](backend/internal/server/router.go)
- [Server Start](backend/commands/start.go)
- [Env Vars](backend/internal/config/config.go)