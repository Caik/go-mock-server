## Build frontend ##
FROM node:20-alpine AS web-builder
WORKDIR /app/web
COPY web/package*.json ./
# Upgrade npm to 11.x (node:20-alpine ships 10.x which has a known bug
# where it silently fails and exits 0 without installing packages).
# NODE_ENV=development prefix applies only to npm ci so devDependencies
# (build tools like Vite, @react-router/dev) are installed.
# npm run build runs as a separate layer and inherits NODE_ENV=production.
RUN npm install -g npm@latest && NODE_ENV=development npm ci
COPY web/ ./
RUN npm run build

## Building binaries ##
FROM golang:1.25-alpine AS builder
RUN apk add --no-cache git make
ARG VERSION=dev
WORKDIR /tmp/go-mock-server
COPY . .
RUN go mod download \
    && LDFLAGS="-s -w -X 'github.com/Caik/go-mock-server/internal/config._version=$VERSION'" \
    && CGO_ENABLED=0 go build -a -installsuffix cgo -o dist/mock-server -ldflags "$LDFLAGS" cmd/mock-server/main.go

## Creating final image ##
FROM alpine:latest
RUN apk add ca-certificates
COPY --from=builder /tmp/go-mock-server/dist/mock-server /app/mock-server
COPY --from=web-builder /app/web/build/client /app/ui
ENTRYPOINT ["/app/mock-server"]
