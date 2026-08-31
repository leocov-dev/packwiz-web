default: help

# build frontend
build-fe:
	cd frontend && npm ci && npm run build

# build backend
build-be:
	cd backend && make build

# build the frontend and the backend
build-all: build-fe build-be

# build the docker image
build-image:
	docker build -t packwiz-web .

# dev env vars for the local postgres started by dev-db-up
DEV_ENV := PWW_MODE=development \
	PWW_PG_PORT=55432 \
	PWW_PG_PASSWORD=insecure-db-password \
	PWW_SESSION_SECRET=insecure-session-secret \
	PWW_ADMIN_PASSWORD=insecure-admin-pass-change-me

# start local postgres for development (docker)
dev-db-up:
	docker compose -f localdev/docker-compose.yml up -d --wait

# stop local dev postgres
dev-db-down:
	docker compose -f localdev/docker-compose.yml down

# run both the frontend and backend in development mode (auto-starts local postgres)
start-dev: dev-db-up
	cd backend && $(DEV_ENV) make start-dev&
	cd frontend && npm run dev

# print help information
help:
	@echo "Available commands:"
	@awk ' \
		/^[a-zA-Z0-9_-]+:/ { \
			if ($$1 != "default:") { \
				cmd = $$1; \
				sub(":", "", cmd); \
				printf "  %-20s %s\n", cmd, last_comment; \
			} \
			last_comment = ""; \
		} \
		/^[#].+/ { \
			sub("^[#] ", "", $$0); \
			last_comment = $$0; \
		} \
	' $(MAKEFILE_LIST)

.PHONY: help