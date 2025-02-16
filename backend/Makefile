BIN_TARGET := $(CURDIR)/bin/backend
BIN_PACKWIZ := $(CURDIR)/bin
GOFMT_FILES?=$$(find . -type f -name '*.go')

default: help

# build the server binary
build:
	go build -o $(BIN_TARGET)

# start the server in development mode (will run build first)
start-dev: build
	PWW_MODE=development $(BIN_TARGET)

# start the server in production mode (will run build first)
start-prod: build
	$(BIN_TARGET)

# run all tests
test:
	@rm -f coverage/coverage.out coverage/coverage.html
	@go test -v -coverpkg=./... -coverprofile=coverage/coverage.out ./...
	@go tool cover -html=coverage/coverage.out -o coverage/coverage.html

# format all code
fmt:
	@gofmt -w $(GOFMT_FILES)

# assert if code is formatted
fmtcheck:
	@bash -c "scripts/gofmtcheck.sh"

# cleanup go.mod files
tidy:
	@go mod tidy

# fetch and install packwiz_data CLI for current arch
fetch-packwiz:
	GOBIN=$(BIN_PACKWIZ) go install github.com/packwiz/packwiz@241f24b550f6fe838913a56bdd58bac2fc53254a

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