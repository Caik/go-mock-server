## Install all dependencies (including dev) ##
FROM --platform=$BUILDPLATFORM node:20-alpine AS web-deps
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci

## Build frontend ##
FROM --platform=$BUILDPLATFORM node:20-alpine AS web-builder
WORKDIR /app/web
COPY web/ ./
COPY --from=web-deps /app/web/node_modules ./node_modules
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
