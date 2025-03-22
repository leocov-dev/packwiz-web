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

# run both the frontend and backend in development mode
start-dev:
	cd backend && make start-dev&
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