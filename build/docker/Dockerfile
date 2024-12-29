## Building binaries ##
FROM golang:1.21-alpine AS builder

ARG TARGETOS=linux
ARG TARGETARCH=amd64

RUN apk add --no-cache git make

WORKDIR /tmp/go-mock-server

COPY . .

RUN go mod download \
    && LDFLAGS="-s -w -X 'github.com/Caik/go-mock-server/internal/config._version=$VERSION'" \
    && CGO_ENABLED=0 GOOS="${TARGETOS}" GOARCH="${TARGETARCH}" go build -a -installsuffix cgo -o dist/mock-server -ldflags "$LDFLAGS" cmd/mock-server/main.go

## Creating final image ##
FROM alpine:latest 

RUN apk add ca-certificates

COPY --from=builder /tmp/go-mock-server/dist/mock-server /app/mock-server

ENTRYPOINT ["/app/mock-server"]