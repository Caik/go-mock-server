# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Run all Go tests
make test
# equivalent: CGO_ENABLED=0 go test -v -count=1 -race -shuffle=on -coverprofile=coverage.txt ./...

# Run a single test
go test ./internal/path/to/package/... -run TestFunctionName

# Build the React admin UI
make build-ui
# equivalent: cd web && npm ci && npm run build

# Format Go code (run before committing)
gofmt -s -w .

# Run locally with Docker
make run_docker

# Build Go binary
go build -v ./...
```

## Architecture

**Go Mock Server** is a dual-server HTTP mocking tool. It runs two servers:
- **Mock Server** (port 8080): Intercepts requests and returns configured mock responses
- **Admin Server** (port 9090): REST API for managing mocks + embedded React admin UI

### Entry Point & DI

`cmd/mock-server/main.go` uses **uber/dig** for dependency injection. All components are wired in `internal/ci/ci.go`. CLI arguments are parsed in `internal/config/app_arguments.go` — key flags: `--mocks-directory` (required), `--port`, `--admin-port`, `--disable-cache`, `--disable-latency`, `--disable-cors`.

### Request Pipeline

When a mock request hits port 8080, it flows through a service pipeline built in `internal/service/mock/mock_factory.go`:

1. **Host resolution** — matches request host/URI to a mock file
2. **Content retrieval** — reads mock file from disk (`internal/service/content/`)
3. **Cache** — optional in-memory caching (disable with `--disable-cache`)
4. **Latency simulation** — adds artificial delay (disable with `--disable-latency`)
5. **Status simulation** — injects error status codes per config
6. **Content-Type** — sets response content type
7. **CORS** — injects CORS headers (disable with `--disable-cors`)

### Mock File Convention

Mock files live in `--mocks-directory` and follow the naming pattern:
```
{host}/{uri}.{method}.{status-code}
```
Examples: `api.example.com/users/GET.200`, `api.example.com/orders/POST.201`

### Key Packages

| Package | Role |
|---|---|
| `internal/server/` | Gin engine setup, middleware, graceful shutdown |
| `internal/server/controller/` | HTTP handlers for both servers |
| `internal/service/mock/` | Mock resolution pipeline (factory + per-concern services) |
| `internal/service/admin/` | CRUD for dynamic mocks and host config |
| `internal/service/content/` | File system I/O for mock files |
| `internal/service/traffic/` | Ring-buffer request log |
| `internal/util/` | Broadcaster (real-time updates), ring buffer, validators |

### Frontend

`web/` is a React 19 + React Router 7 + TypeScript + Tailwind CSS SPA built with Vite. It serves from the admin server via `internal/server/controller/ui_controller.go`. Run `make build-ui` when making frontend changes — the built assets are embedded into the binary.

### Patterns to Follow

- New services should be registered via uber/dig in `internal/ci/ci.go`
- Controllers use Gin's context; all route registration is in `internal/server/controller/controllers.go`
- Use `rs/zerolog` for structured logging (not `log` package)
- Traffic logging uses `internal/util/ring_buffer.go` — bounded, no unbounded growth
- Real-time UI updates use `internal/util/broadcaster.go`

## PR Checklist

Per `.github/pull_request_template.md`:
- Run `gofmt -s -w .` before committing
- Run `make test` before committing
- Update Swagger docs in `docs/` for any API changes
- Update README.md for user-facing changes
- Run `make build-ui` if `web/` was modified
