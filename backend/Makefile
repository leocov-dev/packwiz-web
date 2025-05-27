BIN_TARGET := $(CURDIR)/bin/backend
GOFMT_FILES?=$$(find . -type f -name '*.go')

default: help

# build the server binary
build: fmt
	go build -o $(BIN_TARGET)

# start the server in development mode (will run build first)
start-dev: build
	PWW_MODE=development $(BIN_TARGET)

# start the server in production mode (will run build first)
start-prod: build
	$(BIN_TARGET)

# run all tests
test:
	@mkdir -p coverage
	@rm -f coverage/coverage.out coverage/coverage.html
	@go test -v -coverpkg=./... -coverprofile=coverage/coverage.out ./...
	@go tool cover -html=coverage/coverage.out -o coverage/coverage.html

# format all code
fmt: tidy
	@gofmt -w $(GOFMT_FILES)

# assert if code is formatted
fmtcheck:
	@bash -c "scripts/gofmtcheck.sh"

# cleanup go.mod files
tidy:
	@go mod tidy

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