## Building binaries ##
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /tmp/go-mock-server

COPY . .

RUN go mod download
RUN make build

## Creating final image ##
FROM alpine:latest 

RUN apk add ca-certificates

COPY --from=builder /tmp/go-mock-server/dist/mock-server /app/mock-server

ENTRYPOINT ["/app/mock-server"]