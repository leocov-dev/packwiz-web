FROM node:22.0.0 as frontend

WORKDIR /frontend

COPY ./frontend/package*.json ./

RUN npm ci

COPY ./frontend .

RUN npx vite build \
    --outDir ./dist \
    --mode production


FROM golang:1.23-bookworm as backend

ARG VERSION_TAG
ARG CF_API_KEY

WORKDIR /backend

COPY ./backend .
COPY --from=frontend /frontend/dist ./public/frontend

# this is broken out from call to `make` to improve docker caching
COPY /backend/go.mod /backend/go.sum ./
RUN go mod download && go mod verify

RUN  go build \
     -o ./bin/backend \
     --ldflags="-X 'packwiz-web/main.VersionTag=$VERSION_TAG' -X 'github.com/packwiz-nxt/main.CfApiKey=$CF_API_KEY'"

FROM debian:bookworm-slim as runtime

WORKDIR /app

COPY --from=backend \
    /backend/bin/backend \
    /backend/bin/packwiz \
    /app/

ENTRYPOINT ["/app/backend"]

CMD ["start"]
