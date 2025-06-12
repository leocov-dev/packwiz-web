FROM node:22.0.0 AS frontend

WORKDIR /frontend

COPY ./frontend/package*.json ./

RUN npm ci

COPY ./frontend .

RUN npx vite build \
    --outDir ./dist \
    --mode production


FROM golang:1.23-bookworm AS backend

ARG VERSION_TAG
ARG CF_API_KEY

WORKDIR /backend

# this is broken out from call to `make` to improve docker caching
COPY /backend/go.mod /backend/go.sum ./
RUN go mod download && go mod verify

COPY ./backend .
COPY --from=frontend /frontend/dist ./public/frontend

RUN go build \
     -o ./bin/backend \
     -trimpath \
     --ldflags="\
        -s \
        -w \
        -X 'github.com/packwiz-web/main.VersionTag=$VERSION_TAG' \
        -X 'github.com/packwiz-nxt/main.CfApiKey=$CF_API_KEY' \
     "

FROM debian:bookworm-slim AS runtime

RUN apt-get update && \
    apt-get install -y curl && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=backend \
    /backend/bin/backend \
    /app/

ENTRYPOINT ["/app/backend"]

CMD ["start", "--migrate"]
