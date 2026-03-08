## Build frontend ##
FROM node:20-alpine AS web-builder
WORKDIR /app/web
COPY web/package*.json ./
# NODE_ENV=development prefix applies only to npm ci, ensuring devDependencies
# (build tools like Vite, @react-router/dev) are installed.
# npm run build runs as a separate layer and inherits NODE_ENV=production
# (set by the base image), so the output is a production-optimised bundle.
RUN NODE_ENV=development npm ci
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
