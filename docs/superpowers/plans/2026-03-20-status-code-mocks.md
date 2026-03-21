# Status-Code Specific Mocks Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add status-code suffixes to mock filenames, rename "error simulation" to "status simulation" everywhere, and update the UI to show/manage status codes per mock.

**Architecture:** `StatusSimulationMockService` (renamed from `ErrorMockService`) sets `MockResponse.StatusCode` early in the chain (defaulting to 200). `ContentMockService` reads the status code and performs tiered file lookup: `uri.method.{status}` → host `_default.{status}` → empty body. Cache keys include status code. All "error" terminology is renamed to "status" across Go, API routes, CLI flags, React UI, and docs.

**Tech Stack:** Go 1.25 (Gin, zerolog, go-arg), React 19 + TypeScript (Tailwind CSS, Vite, React Router 7)

**Spec:** `docs/superpowers/specs/2026-03-20-status-code-mocks-design.md`

---

## File Map

### Modified — Go backend
- `internal/service/mock/error_mock_service.go` → rename to `status_simulation_mock_service.go`, update all types/methods
- `internal/service/mock/error_mock_service_test.go` → rename to `status_simulation_mock_service_test.go`
- `internal/service/mock/content_mock_service.go` → update to use status code for tiered file lookup
- `internal/service/mock/content_mock_service_test.go` → update tests for new lookup logic
- `internal/service/mock/mock_service.go` → rename `activeErrorConfig` field to `activeStatusConfig`; update `GenerateCacheKey` to include status code
- `internal/service/mock/mock_factory.go` → rename `disableErrors`/`newErrorMockService` to `disableStatusSimulation`/`newStatusSimulationMockService`
- `internal/service/mock/mock_factory_test.go` → update references
- `internal/service/mock/metadata.go` → rename `MetadataSimulatedError` and `MetadataErrorRuleScope` constants
- `internal/service/content/content_service.go` → add `StatusCode` param to `GetContent`; add `StatusCode` field to `ContentData`
- `internal/service/content/filesystem_content_service.go` → update `GetContent`, `getFinalFilePath`, `filePathToContentData` for `.{status}` suffix
- `internal/config/hosts_config.go` → rename `ErrorsConfig`→`StatusesConfig`, `ErrorConfig`→`StatusConfig`, all methods and validation; expand status code validation from 400–599 to 100–599
- `internal/config/hosts_config_test.go` → update all references
- `internal/config/hosts_config_validation_test.go` → update validation tests
- `internal/config/app_arguments.go` → rename `DisableError` to `DisableStatusSimulation`, update help text
- `internal/server/controller/controllers.go` → rename API routes: `/:host/errors` → `/:host/statuses`, `/:host/errors/:error` → `/:host/statuses/:status`
- `internal/server/controller/admin_hosts_controller.go` → rename `ErrorConfig`→`StatusConfig` fields, handler methods, validation messages
- `internal/server/controller/admin_hosts_controller_test.go` → update references
- `internal/server/controller/admin_mocks_controller.go` → add `Status` header (`x-mock-status`) to `AddDeleteMockRequest`; pass status to service
- `internal/server/controller/admin_mocks_controller_test.go` → update tests
- `internal/service/admin/mock_admin_service.go` → add `Status` to `MockAddDeleteRequest`, `MockListItem`; update `generateMockID`/`decodeMockID` to include status (4 parts)
- `internal/service/admin/mock_admin_service_test.go` → update tests
- `internal/service/admin/host_config_admin_service.go` → rename `ErrorConfig`→`StatusConfig`, `AddUpdateHostErrors`→`AddUpdateHostStatuses`, `DeleteHostError`→`DeleteHostStatus`
- `internal/service/admin/host_config_admin_service_test.go` → update references

### Modified — React frontend
- `web/app/types/host.ts` → rename `ErrorConfig`→`StatusConfig`, `errors`→`statuses` fields in `HostConfig`/`UriConfig`
- `web/app/types/mock.ts` → add `statusCode: number` to `MockDefinition`
- `web/app/services/host.service.ts` → rename API types/mappers/payloads from `error`→`status`; update API URL from `/errors` to `/statuses`
- `web/app/services/mock.service.ts` → add `statusCode` to `MockData`; send `x-mock-status` header; update `toMockDefinition`; update `ApiMock` type
- `web/app/components/ui/MockEditModal.tsx` → add `statusCode` field (number input, default 200); add to `MockFormData`
- `web/app/components/ui/HostEditModal.tsx` → rename "Errors" section to "Status Simulation"; expand status code validation to 100–599; rename `ErrorRow`→`StatusRow`, `ErrorRowError`→`StatusRowError`, etc.
- `web/app/components/details/HostDetail.tsx` → rename "Global Errors" → "Status Simulation"; update field references
- `web/app/components/details/LogDetail.tsx` → no code change needed (renders metadata keys dynamically); metadata key rename in backend handles this
- `web/app/routes/mocks.tsx` → add `statusCode` column to table; pass `statusCode` through mock save/update flows

### Modified — Docs/config
- `README.md` → rename error simulation references; document breaking change; migration instructions
- `docs/` Swagger spec files (if any) → update route descriptions, field names
- `roadmap.md` → update any error simulation references
- `sample-mocks/` → rename sample files to include `.200` suffix

---

## Task 1: Rename config structs and validation

**Files:**
- Modify: `internal/config/hosts_config.go`
- Modify: `internal/config/hosts_config_test.go`
- Modify: `internal/config/hosts_config_validation_test.go`
- Modify: `internal/config/app_arguments.go`

- [ ] **Step 1: Write failing test for new struct names**

In `internal/config/hosts_config_test.go`, add a test that references `StatusesConfig` and `StatusConfig` field names:

```go
func TestHostConfig_StatusesConfig(t *testing.T) {
    p := 50
    cfg := config.HostConfig{
        StatusesConfig: map[string]config.StatusConfig{
            "500": {Percentage: &p},
        },
    }
    if err := cfg.Validate(); err != nil {
        t.Errorf("unexpected error: %v", err)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /Users/cseverino/workspaces/go-mock-server && go test ./internal/config/... 2>&1 | head -30
```

Expected: compile error — `StatusesConfig` undefined.

- [ ] **Step 3: Rename structs and methods in hosts_config.go**

In `internal/config/hosts_config.go`:
- `ErrorsConfig map[string]ErrorConfig` → `StatusesConfig map[string]StatusConfig` (in `HostConfig`)
- `ErrorsConfig map[string]ErrorConfig` → `StatusesConfig map[string]StatusConfig` (in `UriConfig`)
- `type ErrorConfig struct` → `type StatusConfig struct`
- JSON tag `"errors"` → `"statuses"` on `HostConfig.StatusesConfig` and `UriConfig.StatusesConfig`
- `UpdateHostErrorsConfig` → `UpdateHostStatusesConfig`, update field references inside
- `DeleteHostErrorConfig` → `DeleteHostStatusConfig`, update field references inside
- `GetAppropriateErrorsConfig` → `GetAppropriateStatusesConfig`, return type `*map[string]StatusConfig`, update references inside
- Status code validation in `HostConfig.Validate()`: change `intErrorCode < 400` to `intErrorCode < 100`; update error message: `"error should belong to either 4xx or 5xx classes"` → `"status code must be between 100 and 599"`
- Status code validation in `UriConfig.validate()`: change `intStatusCode < 400` to `intStatusCode < 100`; update error message: `"error status code should be between 400 and 599"` → `"status code must be between 100 and 599"`
- `UriConfig.validate()` nil check: change `if u.ErrorsConfig == nil && u.LatencyConfig == nil` and update error message from `"latency or errors should not be both null"` → `"latency or statuses should not be both null"`
- `ErrorConfig.validate()` → `StatusConfig.validate()`, update all error messages to say "status config" instead of "error config"

- [ ] **Step 4: Update app_arguments.go**

```go
// Change:
DisableError bool `arg:"--disable-error" help:"disable error simulation"`
// To:
DisableStatusSimulation bool `arg:"--disable-status-simulation" help:"disable status simulation"`
```

- [ ] **Step 5: Update all test files to use new names**

In `hosts_config_test.go` and `hosts_config_validation_test.go`:
- Replace all `ErrorsConfig` → `StatusesConfig`
- Replace all `ErrorConfig` → `StatusConfig`
- Replace all `GetAppropriateErrorsConfig` → `GetAppropriateStatusesConfig`

- [ ] **Step 6: Run tests**

```bash
cd /Users/cseverino/workspaces/go-mock-server && go test ./internal/config/... -v 2>&1 | tail -30
```

Expected: all tests pass.

- [ ] **Step 7: Commit**

```bash
cd /Users/cseverino/workspaces/go-mock-server
git add internal/config/
git commit -m "refactor: rename ErrorConfig to StatusConfig and expand status code validation to 1xx-5xx"
```

---

## Task 2: Rename metadata constants and StatusSimulationMockService

**Files:**
- Modify: `internal/service/mock/metadata.go`
- Rename + modify: `internal/service/mock/error_mock_service.go` → `status_simulation_mock_service.go`
- Rename + modify: `internal/service/mock/error_mock_service_test.go` → `status_simulation_mock_service_test.go`
- Modify: `internal/service/mock/mock_service.go`
- Modify: `internal/service/mock/mock_factory.go`
- Modify: `internal/service/mock/mock_factory_test.go`

- [ ] **Step 1: Update metadata.go**

```go
// Change:
MetadataSimulatedError   = "Simulated Error"
MetadataErrorRuleScope   = "Error Rule Scope"
// To:
MetadataSimulatedStatus  = "Simulated Status"
MetadataStatusRuleScope  = "Status Rule Scope"
```

- [ ] **Step 2: Create status_simulation_mock_service.go from error_mock_service.go**

Copy `error_mock_service.go` to `status_simulation_mock_service.go` and apply these changes:
- `type errorMockService struct` → `type statusSimulationMockService struct`
- `type errorPercentageWrapper struct` → `type statusPercentageWrapper struct`; rename field `originalErrorConfig config.ErrorConfig` → `originalStatusConfig config.StatusConfig`
- `hostsConfig.GetAppropriateErrorsConfig` → `hostsConfig.GetAppropriateStatusesConfig`
- `e.drawError` → `e.drawStatus`; `drawError` method → `drawStatus`; update all internal references
- `newErrorMockService` → `newStatusSimulationMockService`
- Log message: `"simulating error"` → `"simulating status"`
- `resp.activeErrorConfig` → `resp.activeStatusConfig`
- `resp.AddMetadata(MetadataSimulatedError, "true")` → `resp.AddMetadata(MetadataSimulatedStatus, "true")`
- `resp.AddMetadata(MetadataErrorRuleScope, scope)` → `resp.AddMetadata(MetadataStatusRuleScope, scope)`
- Delete `error_mock_service.go`

- [ ] **Step 3: Write failing test for cache key including status code**

In `internal/service/mock/mock_service_test.go` (or create it), add:

```go
func TestGenerateCacheKey_IncludesStatusCode(t *testing.T) {
    req200 := MockRequest{Host: "example.com", Method: "GET", URI: "/api/users", StatusCode: 200}
    req500 := MockRequest{Host: "example.com", Method: "GET", URI: "/api/users", StatusCode: 500}
    if GenerateCacheKey(req200) == GenerateCacheKey(req500) {
        t.Error("cache keys for different status codes must differ")
    }
}
```

Run: `go test ./internal/service/mock/... -run TestGenerateCacheKey_IncludesStatusCode`
Expected: FAIL (cache key does not include status code yet).

- [ ] **Step 4: Update mock_service.go**

```go
// Change:
activeErrorConfig   *config.ErrorConfig
// To:
activeStatusConfig  *config.StatusConfig
```

Also update `GenerateCacheKey` to include status code (this requires status code to be on `MockRequest`):

```go
// Add StatusCode to MockRequest:
type MockRequest struct {
    Host       string
    Method     string
    URI        string
    Accept     string
    Uuid       string
    StatusCode int // set by StatusSimulationMockService; 0 means not yet determined
}

// Update GenerateCacheKey:
func GenerateCacheKey(mockRequest MockRequest) string {
    return strings.Join([]string{
        mockRequest.Host,
        mockRequest.Method,
        mockRequest.URI,
        strconv.Itoa(mockRequest.StatusCode),
    }, ":")
}
```

Import `"strconv"` in mock_service.go.

- [ ] **Step 5: Update StatusSimulationMockService to set StatusCode on MockRequest**

**Important behavioral change:** Unlike the old `errorMockService` which short-circuited (returned early with empty body when an error was drawn and never called downstream), the new `statusSimulationMockService` ALWAYS calls the downstream chain — including `ContentMockService` — so that status-specific mock files (e.g. `api/users.get.500`) can be served. The `StatusCode` is set on `MockRequest` before passing downstream, and after the response returns, the drawn status code is applied to the response along with the simulation metadata.

In `status_simulation_mock_service.go`, replace the entire `getMockResponse` method with:

```go
func (e *statusSimulationMockService) getMockResponse(mockRequest MockRequest) *MockResponse {
    statusesConfig, scope := e.hostsConfig.GetAppropriateStatusesConfig(mockRequest.Host, mockRequest.URI)

    statusCode := 200
    var drawnWrapper *statusPercentageWrapper

    if statusesConfig != nil {
        drawnWrapper = e.drawStatus(statusesConfig)
        if drawnWrapper != nil {
            statusCode = drawnWrapper.statusCode
        }
    }

    // Always set status code on request before passing downstream,
    // so ContentMockService can find the correct status-specific file.
    mockRequest.StatusCode = statusCode

    log.Info().
        Str("uuid", mockRequest.Uuid).
        Int("status_code", statusCode).
        Msg("simulating status")

    resp := e.nextOrNil(mockRequest)

    if drawnWrapper != nil {
        if resp == nil {
            emptyResponse := []byte("")
            resp = &MockResponse{Data: &emptyResponse}
        }
        resp.StatusCode = statusCode
        resp.activeStatusConfig = &drawnWrapper.originalStatusConfig
        resp.AddMetadata(MetadataSimulatedStatus, "true")
        resp.AddMetadata(MetadataStatusRuleScope, scope)
    }

    return resp
}
```

Note: the `log.Info` for "simulating status" is now called unconditionally (even for `statusCode=200`) when a config is present and a status was drawn. This replaces the old log message "simulating error".

- [ ] **Step 6: Update mock_factory.go**

```go
// Change parameter name:
disableErrors bool → disableStatusSimulation bool

// Change instantiation:
if !disableErrors {
    addNextFn(newErrorMockService(hostsConfig))
}
// To:
if !disableStatusSimulation {
    addNextFn(newStatusSimulationMockService(hostsConfig))
}

// Change in initServiceChain signature and NewMockServiceFactory:
arguments.DisableError → arguments.DisableStatusSimulation

// Update log message:
"error while starting HostResolutionMockService" (keep as-is, not related)
```

- [ ] **Step 7: Create status_simulation_mock_service_test.go**

Copy `error_mock_service_test.go` to `status_simulation_mock_service_test.go` and rename:
- `TestErrorMockService_*` → `TestStatusSimulationMockService_*`
- `newErrorMockService` → `newStatusSimulationMockService`
- `errorMockService` → `statusSimulationMockService`
- `ErrorsConfig` → `StatusesConfig`
- `ErrorConfig` → `StatusConfig`
- `drawError` → `drawStatus`
- Delete `error_mock_service_test.go`

Also update mock_factory_test.go to use new names.

- [ ] **Step 8: Run tests**

```bash
cd /Users/cseverino/workspaces/go-mock-server && go test ./internal/service/mock/... -v 2>&1 | tail -40
```

Expected: all tests pass.

- [ ] **Step 9: Commit**

```bash
cd /Users/cseverino/workspaces/go-mock-server
git add internal/service/mock/
git commit -m "refactor: rename ErrorMockService to StatusSimulationMockService and update metadata keys"
```

---

## Task 3: Update ContentService interface and filesystem implementation for status-code suffix

**Files:**
- Modify: `internal/service/content/content_service.go`
- Modify: `internal/service/content/filesystem_content_service.go`
- Modify (tests): any files testing content service

- [ ] **Step 1: Write a failing test for status-code file lookup**

In a new file `internal/service/content/filesystem_content_service_status_test.go` (or existing test file if one exists):

```go
func TestFilesystemContentService_GetContent_StatusSuffix(t *testing.T) {
    t.Run("finds file with status suffix", func(t *testing.T) {
        dir := t.TempDir()
        hostDir := filepath.Join(dir, "example.com")
        os.MkdirAll(hostDir, 0755)
        os.WriteFile(filepath.Join(hostDir, "api.users.get.200"), []byte("ok"), 0644)

        svc := NewFilesystemContentService(&config.MocksDirectoryConfig{Path: dir})
        result, err := svc.GetContent("example.com", "/api/users", "GET", "test", 200)

        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        if string(*result.Data) != "ok" {
            t.Errorf("expected 'ok', got %q", string(*result.Data))
        }
    })

    t.Run("returns empty body when no file found for non-200 status", func(t *testing.T) {
        dir := t.TempDir()
        hostDir := filepath.Join(dir, "example.com")
        os.MkdirAll(hostDir, 0755)

        svc := NewFilesystemContentService(&config.MocksDirectoryConfig{Path: dir})
        result, err := svc.GetContent("example.com", "/api/users", "GET", "test", 500)

        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        if result.Data == nil || len(*result.Data) != 0 {
            t.Error("expected empty body")
        }
    })

    t.Run("falls back to _default.{status} file for non-200 status", func(t *testing.T) {
        dir := t.TempDir()
        hostDir := filepath.Join(dir, "example.com")
        os.MkdirAll(hostDir, 0755)
        os.WriteFile(filepath.Join(hostDir, "_default.500"), []byte("default 500"), 0644)

        svc := NewFilesystemContentService(&config.MocksDirectoryConfig{Path: dir})
        result, err := svc.GetContent("example.com", "/api/users", "GET", "test", 500)

        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        if string(*result.Data) != "default 500" {
            t.Errorf("expected 'default 500', got %q", string(*result.Data))
        }
    })
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /Users/cseverino/workspaces/go-mock-server && go test ./internal/service/content/... 2>&1 | head -20
```

Expected: compile error — `GetContent` signature mismatch.

- [ ] **Step 3: Update ContentService interface**

In `internal/service/content/content_service.go`:

```go
type ContentService interface {
    GetContent(host, uri, method, uuid string, statusCode int) (*ContentResult, error)
    SetContent(host, uri, method, uuid string, statusCode int, data *[]byte) error
    DeleteContent(host, uri, method, uuid string, statusCode int) error
    ListContents(uuid string) (*[]ContentData, error)
    Subscribe(subscriberId string, eventTypes ...ContentEventType) <-chan ContentEvent
    Unsubscribe(subscriberId string)
}
```

Also add `StatusCode int` to `ContentData`:

```go
type ContentData struct {
    Host       string
    Uri        string
    Method     string
    StatusCode int
}
```

- [ ] **Step 4: Update getFinalFilePath in filesystem_content_service.go**

Add `statusCode int` parameter and append `.{statusCode}` to the file path:

```go
func (f *FilesystemContentService) getFinalFilePath(host, uri, method string, statusCode int) (string, error) {
    // ... existing validation unchanged ...

    // After building finalPath and appending method:
    finalPath += "." + strings.ToLower(method) + "." + strconv.Itoa(statusCode)

    // ... existing path traversal check unchanged ...
}
```

Import `"strconv"`.

- [ ] **Step 5: Update GetContent in filesystem_content_service.go**

```go
func (f *FilesystemContentService) GetContent(host, uri, method, uuid string, statusCode int) (*ContentResult, error) {
    absolutePath, err := f.getFinalFilePath(host, uri, method, statusCode)
    if err != nil {
        return nil, err
    }

    data, err := os.ReadFile(absolutePath)
    if err != nil {
        // For non-200: try _default.{statusCode}, then return empty body
        if statusCode != 200 {
            defaultPath, defErr := f.getDefaultFilePath(host, statusCode)
            if defErr == nil {
                defaultData, defReadErr := os.ReadFile(defaultPath)
                if defReadErr == nil {
                    return &ContentResult{
                        Data:   &defaultData,
                        Source: "filesystem",
                        Path:   defaultPath,
                    }, nil
                }
            }
            // No default file found — return empty body (not an error)
            empty := []byte("")
            return &ContentResult{
                Data:   &empty,
                Source: "filesystem",
                Path:   "",
            }, nil
        }

        log.Info().Str("uuid", uuid).Str("path", absolutePath).Msg("mock not found")
        return nil, errors.New("mock not found")
    }

    return &ContentResult{
        Data:   &data,
        Source: "filesystem",
        Path:   absolutePath,
    }, nil
}

func (f *FilesystemContentService) getDefaultFilePath(host string, statusCode int) (string, error) {
    if !util.HostRegex.MatchString(host) {
        return "", errors.New("invalid host")
    }
    path := filepath.Join(
        strings.TrimSuffix(f.mocksDirConfig.Path, pathSeparator),
        host,
        "_default."+strconv.Itoa(statusCode),
    )
    mocksDir := filepath.Clean(f.mocksDirConfig.Path)
    rel, err := filepath.Rel(mocksDir, filepath.Clean(path))
    if err != nil || strings.HasPrefix(rel, "..") {
        return "", errors.New("invalid path")
    }
    return filepath.Join(mocksDir, rel), nil
}
```

- [ ] **Step 6: Update SetContent and DeleteContent**

Only the signatures and the first `getFinalFilePath` call change. Copy the existing method bodies verbatim after the `getFinalFilePath` call — they are unchanged:

```go
func (f *FilesystemContentService) SetContent(host, uri, method, uuid string, statusCode int, data *[]byte) error {
    absolutePath, err := f.getFinalFilePath(host, uri, method, statusCode)

    if err != nil {
        return err
    }

    // making sure all parent dirs are created
    parentDir := absolutePath[:strings.LastIndex(absolutePath, pathSeparator)+1]
    err = os.MkdirAll(parentDir, os.ModePerm)

    if err != nil {
        msg := fmt.Sprintf("error while creating parent directories: %v", err)
        log.Err(err).Stack().Str("uuid", uuid).Str("path", absolutePath).Msg("error while creating parent directories")
        return errors.New(msg)
    }

    err = os.WriteFile(absolutePath, *data, 0644)

    if err != nil {
        msg := fmt.Sprintf("error while writing file: %v", err)
        log.Err(err).Stack().Str("uuid", uuid).Str("path", absolutePath).Msg("error while writing file")
        return errors.New(msg)
    }

    return nil
}

func (f *FilesystemContentService) DeleteContent(host, uri, method, uuid string, statusCode int) error {
    absolutePath, err := f.getFinalFilePath(host, uri, method, statusCode)

    if err != nil {
        return err
    }

    if err := os.Remove(absolutePath); err != nil {
        msg := fmt.Sprintf("error while removing file: %v", err)
        log.Err(err).Stack().Str("uuid", uuid).Str("path", absolutePath).Msg("error while removing file")
        return errors.New(msg)
    }

    return nil
}
```

- [ ] **Step 7: Update filePathToContentData for new naming convention**

The file path now has format: `host/uri.method.status`. The existing code finds the last dot to split method from the rest. With the new format there are two trailing dots — the second-to-last separates method from uri, the last separates status from method.

Also: `_default.{status}` files must be skipped during listing (they are fallback files, not real mocks).

Root-path handling note: for root paths, the file is named e.g. `root.get.200` (the existing `rootToken` logic appends "root" before the method). With two trailing dots, we need to find the method boundary as `secondLastDotIndex` and ensure the root suffix trimming still works on the URI segment (e.g. `/root` → `/`).

Replace the existing `filePathToContentData` body with:

```go
func (f *FilesystemContentService) filePathToContentData(path string) (*ContentData, error) {
    rootPath := strings.TrimSuffix(f.mocksDirConfig.Path, pathSeparator) + pathSeparator
    relativePath := strings.TrimPrefix(path, rootPath)

    firstSlashIndex := strings.Index(relativePath, pathSeparator)

    // Skip _default.* files — these are fallbacks, not real mocks
    if firstSlashIndex != -1 {
        fileName := relativePath[firstSlashIndex+1:]
        if strings.HasPrefix(fileName, "_default.") {
            return nil, fmt.Errorf("skipping fallback file: %s", path)
        }
    }

    // Expect format: host/uri.method.status — two trailing dots
    lastDotIndex := strings.LastIndex(relativePath, ".")
    if lastDotIndex == -1 {
        return nil, fmt.Errorf("incorrect file name pattern, ignoring it: %s", path)
    }
    secondLastDotIndex := strings.LastIndex(relativePath[:lastDotIndex], ".")

    if firstSlashIndex == -1 || secondLastDotIndex == -1 || firstSlashIndex >= secondLastDotIndex {
        return nil, fmt.Errorf("incorrect file name pattern, ignoring it: %s", path)
    }

    host := relativePath[:firstSlashIndex]
    uri := relativePath[firstSlashIndex:secondLastDotIndex]
    method := strings.ToUpper(relativePath[secondLastDotIndex+1 : lastDotIndex])
    statusStr := relativePath[lastDotIndex+1:]

    statusCode, err := strconv.Atoi(statusStr)
    if err != nil || statusCode < 100 || statusCode > 599 {
        return nil, fmt.Errorf("invalid status code in filename: %s", path)
    }

    // validating host
    if !util.HostRegex.MatchString(host) {
        return nil, fmt.Errorf("invalid host: %s", host)
    }

    // validating URI
    if !util.UriRegex.MatchString(uri) {
        return nil, fmt.Errorf("invalid uri: %s", uri)
    }

    // validating method
    if !util.HttpMethodRegex.MatchString(method) {
        return nil, fmt.Errorf("invalid method: %s", method)
    }

    // checking if root suffix has been added (e.g. uri ends with /root → trim to /)
    if strings.HasSuffix(uri, fmt.Sprintf("%s%s", pathSeparator, rootToken)) {
        uri = strings.TrimSuffix(uri, rootToken)
    }

    return &ContentData{
        Host:       host,
        Uri:        uri,
        Method:     method,
        StatusCode: statusCode,
    }, nil
}
```

- [ ] **Step 8: Update ContentMockService to pass status code**

In `internal/service/mock/content_mock_service.go`:

```go
func (c *contentMockService) getMockResponse(mockRequest MockRequest) *MockResponse {
    result, err := c.contentService.GetContent(
        mockRequest.Host, mockRequest.URI, mockRequest.Method, mockRequest.Uuid,
        mockRequest.StatusCode,
    )

    if err != nil {
        if errors.Is(err, errContentServiceNotFound) {
            return c.new500Response(err)
        }
        // 200 file not found
        empty := []byte("")
        resp := &MockResponse{
            StatusCode: mockRequest.StatusCode,
            Data:       &empty,
        }
        resp.AddMetadata(MetadataMatched, "false")
        return resp
    }

    statusCode := mockRequest.StatusCode
    if statusCode == 0 {
        statusCode = 200
    }

    resp := &MockResponse{
        StatusCode: statusCode,
        Data:       result.Data,
    }
    resp.AddMetadata(MetadataMatched, "true")
    resp.AddMetadata(MetadataSource, result.Source)
    resp.AddMetadata(MetadataPath, result.Path)
    return resp
}
```

Remove the `new404Response` method (no longer used).

- [ ] **Step 9: Temporarily fix all callers of old ContentService methods to compile**

The interface change forces all callers to add a `statusCode int` argument immediately. Use a placeholder value of `0` for now; Task 4 will replace it with the real decoded value.

In `internal/service/admin/mock_admin_service.go`, temporarily update calls:
```go
// GetMockContent — placeholder 0; replaced in Task 4
result, err := m.contentService.GetContent(host, uri, method, uuid, 0)

// AddUpdateMock — placeholder 0; replaced in Task 4
return m.contentService.SetContent(addRequest.Host, addRequest.URI, addRequest.Method, uuid, 0, addRequest.Data)

// DeleteMock — placeholder 0; replaced in Task 4
return m.contentService.DeleteContent(addRequest.Host, addRequest.URI, addRequest.Method, uuid, 0)

// DeleteMockByID — placeholder 0; replaced in Task 4
return m.contentService.DeleteContent(host, uri, method, uuid, 0)
```

Also update the `mockContentService` test stub in `content_mock_service_test.go` and any other test files to match the new interface signatures.

- [ ] **Step 10: Run tests**

```bash
cd /Users/cseverino/workspaces/go-mock-server && go test ./internal/service/content/... -v 2>&1 | tail -40
```

Expected: all tests pass.

- [ ] **Step 11: Commit**

```bash
cd /Users/cseverino/workspaces/go-mock-server
git add internal/service/content/ internal/service/mock/content_mock_service.go internal/service/mock/content_mock_service_test.go
git commit -m "feat: add status-code suffix to mock file resolution with _default fallback"
```

---

## Task 4: Update admin service — mock ID encoding and status code

**Files:**
- Modify: `internal/service/admin/mock_admin_service.go`
- Modify: `internal/service/admin/mock_admin_service_test.go`

- [ ] **Step 1: Write failing tests**

In `mock_admin_service_test.go`, add:

```go
func TestGenerateMockID_IncludesStatus(t *testing.T) {
    id := generateMockID("example.com", "/api/users", "GET", 200)
    host, uri, method, status, err := decodeMockID(id)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if host != "example.com" || uri != "/api/users" || method != "GET" || status != 200 {
        t.Errorf("unexpected decoded values: %s %s %s %d", host, uri, method, status)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /Users/cseverino/workspaces/go-mock-server && go test ./internal/service/admin/... 2>&1 | head -20
```

Expected: compile error.

- [ ] **Step 3: Update mock_admin_service.go**

```go
type MockAddDeleteRequest struct {
    Host       string
    URI        string
    Method     string
    StatusCode int
    Data       *[]byte
}

type MockListItem struct {
    ID         string `json:"id"`
    Host       string `json:"host"`
    URI        string `json:"uri"`
    Method     string `json:"method"`
    StatusCode int    `json:"status_code"`
}

// Update generateMockID to include status:
func generateMockID(host, uri, method string, statusCode int) string {
    data := fmt.Sprintf("%s%s%s%s%s%s%d",
        host, mockIDSeparator, uri, mockIDSeparator, method, mockIDSeparator, statusCode)
    return base64.URLEncoding.EncodeToString([]byte(data))
}

// Update decodeMockID to return status:
func decodeMockID(id string) (host, uri, method string, statusCode int, err error) {
    data, err := base64.URLEncoding.DecodeString(id)
    if err != nil {
        return "", "", "", 0, fmt.Errorf("failed to decode mock ID: %v", err)
    }
    parts := strings.SplitN(string(data), mockIDSeparator, 4)
    if len(parts) != 4 {
        return "", "", "", 0, fmt.Errorf("invalid mock ID format")
    }
    sc, err := strconv.Atoi(parts[3])
    if err != nil {
        return "", "", "", 0, fmt.Errorf("invalid status code in mock ID: %v", err)
    }
    return parts[0], parts[1], parts[2], sc, nil
}

// Update all methods to pass statusCode to contentService:
func (m *MockAdminService) AddUpdateMock(addRequest MockAddDeleteRequest, uuid string) error {
    return m.contentService.SetContent(addRequest.Host, addRequest.URI, addRequest.Method, uuid, addRequest.StatusCode, addRequest.Data)
}

func (m *MockAdminService) DeleteMockByID(id, uuid string) error {
    host, uri, method, statusCode, err := decodeMockID(id)
    // ...
    return m.contentService.DeleteContent(host, uri, method, uuid, statusCode)
}

func (m *MockAdminService) GetMockContent(id, uuid string) ([]byte, error) {
    host, uri, method, statusCode, err := decodeMockID(id)
    // ...
    result, err := m.contentService.GetContent(host, uri, method, uuid, statusCode)
    // ...
}

// Update ListMocks to include status code:
mocks[i] = MockListItem{
    ID:         generateMockID(c.Host, c.Uri, c.Method, c.StatusCode),
    Host:       c.Host,
    URI:        c.Uri,
    Method:     c.Method,
    StatusCode: c.StatusCode,
}
```

- [ ] **Step 4: Update tests**

Update `mock_admin_service_test.go` to use 4-part IDs and pass statusCode.

- [ ] **Step 5: Run tests**

```bash
cd /Users/cseverino/workspaces/go-mock-server && go test ./internal/service/admin/... -v 2>&1 | tail -30
```

Expected: all tests pass.

- [ ] **Step 6: Commit**

```bash
cd /Users/cseverino/workspaces/go-mock-server
git add internal/service/admin/
git commit -m "feat: include status code in mock ID encoding and admin service operations"
```

---

## Task 5: Update admin service — host config (rename errors → statuses)

**Files:**
- Modify: `internal/service/admin/host_config_admin_service.go`
- Modify: `internal/service/admin/host_config_admin_service_test.go`

- [ ] **Step 1: Update host_config_admin_service.go**

Rename throughout:
- `HostAddDeleteRequest.ErrorConfig` → `StatusConfig map[string]config.StatusConfig`
- `AddUpdateHostErrors` → `AddUpdateHostStatuses`
- `DeleteHostError` → `DeleteHostStatus`
- Internal calls: `UpdateHostErrorsConfig` → `UpdateHostStatusesConfig`, `DeleteHostErrorConfig` → `DeleteHostStatusConfig`
- Log messages and error messages: "errors" → "statuses"

```go
type HostAddDeleteRequest struct {
    Host          string
    LatencyConfig *config.LatencyConfig
    StatusConfig  map[string]config.StatusConfig
    UriConfig     map[string]config.UriConfig
}

func (h *HostsConfigAdminService) AddUpdateHost(addRequest HostAddDeleteRequest) (*config.HostConfig, error) {
    hostConfig := config.HostConfig{
        LatencyConfig:  addRequest.LatencyConfig,
        StatusesConfig: addRequest.StatusConfig,
        UrisConfig:     addRequest.UriConfig,
    }
    // ...
}

func (h *HostsConfigAdminService) AddUpdateHostStatuses(req HostAddDeleteRequest) (*config.HostConfig, error) {
    newHostConfig := config.HostConfig{StatusesConfig: req.StatusConfig}
    if err := newHostConfig.Validate(); err != nil {
        return nil, fmt.Errorf("error while validating host statuses config: %v", err)
    }
    return h.hostsConfig.UpdateHostStatusesConfig(req.Host, newHostConfig.StatusesConfig)
}

func (h *HostsConfigAdminService) DeleteHostStatus(host, statusCode string) (*config.HostConfig, error) {
    return h.hostsConfig.DeleteHostStatusConfig(host, statusCode)
}
```

- [ ] **Step 2: Update tests**

In `host_config_admin_service_test.go`: rename all method calls and field names.

- [ ] **Step 3: Run tests**

```bash
cd /Users/cseverino/workspaces/go-mock-server && go test ./internal/service/admin/... -v 2>&1 | tail -30
```

Expected: all tests pass.

- [ ] **Step 4: Commit**

```bash
cd /Users/cseverino/workspaces/go-mock-server
git add internal/service/admin/host_config_admin_service.go internal/service/admin/host_config_admin_service_test.go
git commit -m "refactor: rename host config admin service error methods to status methods"
```

---

## Task 6: Update HTTP controllers

**Files:**
- Modify: `internal/server/controller/controllers.go`
- Modify: `internal/server/controller/admin_hosts_controller.go`
- Modify: `internal/server/controller/admin_hosts_controller_test.go`
- Modify: `internal/server/controller/admin_mocks_controller.go`
- Modify: `internal/server/controller/admin_mocks_controller_test.go`

- [ ] **Step 1: Update controllers.go — rename routes**

```go
// Change:
r.POST("/:host/errors", controller.handleErrorsAddUpdate)
r.DELETE("/:host/errors/:error", controller.handleErrorDelete)
// To:
r.POST("/:host/statuses", controller.handleStatusesAddUpdate)
r.DELETE("/:host/statuses/:status", controller.handleStatusDelete)
```

- [ ] **Step 2: Update admin_hosts_controller.go**

- `AddDeleteGetHostRequest.ErrorConfig` → `StatusConfig map[string]config.StatusConfig`; JSON tag `"errors"` → `"statuses"`
- `handleErrorsAddUpdate` → `handleStatusesAddUpdate`; update log messages, service call to `AddUpdateHostStatuses`, response message
- `handleErrorDelete` → `handleStatusDelete`; `c.Param("error")` → `c.Param("status")`; `errorCode` → `statusCode`; service call to `DeleteHostStatus`
- `validate()` method: `needsErrors` param → `needsStatuses`; error messages updated
- `a.service.AddUpdateHostErrors` → `a.service.AddUpdateHostStatuses`
- `a.service.DeleteHostError` → `a.service.DeleteHostStatus`
- All log messages: "errors config" → "statuses config"

- [ ] **Step 3: Update admin_mocks_controller.go**

Add `Status` to the mock request header binding, and add a parsed integer field to store it after validation:

```go
type AddDeleteMockRequest struct {
    Host           string `header:"x-mock-host" binding:"required"`
    Uri            string `header:"x-mock-uri" binding:"required"`
    Method         string `header:"x-mock-method" binding:"required"`
    StatusCode     string `header:"x-mock-status"`
    statusCodeInt  int    // populated by validate()
}
```

In `validate()`, add status code parsing after the existing method validation:
```go
// Default to 200 if not provided
if r.StatusCode == "" {
    r.StatusCode = "200"
}
sc, err := strconv.Atoi(r.StatusCode)
if err != nil || sc < 100 || sc > 599 {
    return errors.New("invalid status code provided: must be between 100 and 599")
}
r.statusCodeInt = sc
return nil
```

Update **all four handler methods** that call the mock service — `handleMockCreate`, `handleMockAddUpdate`, `handleMockUpdate`, and `handleMockDelete` — to pass `statusCode` to the service:

```go
// In handleMockCreate, handleMockAddUpdate:
err = a.service.AddUpdateMock(admin.MockAddDeleteRequest{
    Host:       req.Host,
    URI:        req.Uri,
    Method:     req.Method,
    StatusCode: req.statusCodeInt,
    Data:       &data,
}, uuid)

// In handleMockDelete:
err := a.service.DeleteMock(admin.MockAddDeleteRequest{
    Host:       addReq.Host,
    URI:        addReq.Uri,
    Method:     addReq.Method,
    StatusCode: addReq.statusCodeInt,
}, uuid)

// In handleMockUpdate (delete old + create new):
if err := a.service.DeleteMockByID(id, uuid); err != nil { ... }
err = a.service.AddUpdateMock(admin.MockAddDeleteRequest{
    Host:       req.Host,
    URI:        req.Uri,
    Method:     req.Method,
    StatusCode: req.statusCodeInt,
    Data:       &data,
}, uuid)
```

Add `"strconv"` import if not already present.

- [ ] **Step 4: Update tests**

Update `admin_hosts_controller_test.go` and `admin_mocks_controller_test.go` to use new field names, routes, and status code header.

- [ ] **Step 5: Run all backend tests**

```bash
cd /Users/cseverino/workspaces/go-mock-server && go test ./... 2>&1 | tail -40
```

Expected: all tests pass.

- [ ] **Step 6: Commit**

```bash
cd /Users/cseverino/workspaces/go-mock-server
git add internal/server/controller/
git commit -m "refactor: rename error routes to status routes and add status code to mock API"
```

---

## Task 7: Update React frontend — types, services, and mock table

**Files:**
- Modify: `web/app/types/host.ts`
- Modify: `web/app/types/mock.ts`
- Modify: `web/app/services/host.service.ts`
- Modify: `web/app/services/mock.service.ts`
- Modify: `web/app/routes/mocks.tsx`

- [ ] **Step 1: Update host.ts types**

```typescript
export interface HostConfig {
  hostname: string;
  latency?: LatencyConfig;
  statuses?: Record<string, StatusConfig>; // renamed from errors
  uris?: Record<string, UriConfig>;
}

export interface StatusConfig {  // renamed from ErrorConfig
  percentage: number;
  latency?: LatencyConfig;
}

export interface UriConfig {
  latency?: LatencyConfig;
  statuses?: Record<string, StatusConfig>; // renamed from errors
}
```

- [ ] **Step 2: Update mock.ts**

```typescript
export interface MockDefinition {
  id: string;
  endpoint: string;
  method: string;
  host: string;
  statusCode: number;
}
```

- [ ] **Step 3: Update host.service.ts**

- Rename all `ApiErrorConfig` → `ApiStatusConfig`; `ApiUriConfig.errors` → `ApiUriConfig.statuses`; `ApiHostConfig.errors` → `ApiHostConfig.statuses`
- Rename mappers: `toErrorsConfig` → `toStatusesConfig`
- Update `toUriConfig`: `api.errors` → `api.statuses`; `errors:` → `statuses:`
- Update `toHostConfig`: `api.errors` → `api.statuses`; `errors:` → `statuses:`
- Update `HostSaveData`: `errors?` → `statuses?`
- Update `UriPayload`: `errors?` → `statuses?`
- Update `saveHost` payload key: `errors:` → `statuses:`

- [ ] **Step 4: Update mock.service.ts**

```typescript
interface ApiMock {
  id: string;
  host: string;
  uri: string;
  method: string;
  status_code: number;
}

function toMockDefinition(item: ApiMock): MockDefinition {
  return {
    id: item.id,
    endpoint: item.uri,
    method: item.method,
    host: item.host,
    statusCode: item.status_code,
  };
}

export interface MockData {
  host: string;
  uri: string;
  method: string;
  statusCode: number;
  body: string;
}

// In createMock, updateMock: add header
'x-mock-status': String(mock.statusCode),

// In deleteMock: add header
'x-mock-status': String(mock.statusCode),
```

- [ ] **Step 5: Update mocks.tsx — add status column and wire statusCode through save/update**

Add `statusCode` column to the table columns definition:
```tsx
{ key: 'statusCode', header: 'Status', render: (mock) => (
  <span className={`status-badge ${getStatusClass(mock.statusCode)}`}>{mock.statusCode}</span>
)}
```

Update `handleSave`:
```tsx
await createMock({
  host: data.host,
  uri: data.endpoint,
  method: data.method,
  statusCode: data.statusCode,
  body: data.responseBody,
});
```

Update `handleUpdate` similarly.

- [ ] **Step 6: Run frontend build to check for type errors**

```bash
cd /Users/cseverino/workspaces/go-mock-server/web && npm run typecheck 2>&1 | tail -30
```

If `typecheck` script doesn't exist:
```bash
cd /Users/cseverino/workspaces/go-mock-server/web && npx tsc --noEmit 2>&1 | tail -30
```

Expected: no type errors.

- [ ] **Step 7: Commit**

```bash
cd /Users/cseverino/workspaces/go-mock-server
git add web/app/types/ web/app/services/ web/app/routes/mocks.tsx
git commit -m "feat: add status code to mock definition and rename error types to status types in frontend"
```

---

## Task 8: Update React frontend — MockEditModal and HostEditModal

**Files:**
- Modify: `web/app/components/ui/MockEditModal.tsx`
- Modify: `web/app/components/ui/HostEditModal.tsx`
- Modify: `web/app/components/details/HostDetail.tsx`

- [ ] **Step 1: Update MockEditModal.tsx**

Add `statusCode: number` to `MockFormData`:
```typescript
export interface MockFormData {
  host: string;
  endpoint: string;
  method: HttpMethod;
  statusCode: number;
  responseBody: string;
}
```

Add status code field to initial state (default 200):
```typescript
const [formData, setFormData] = useState<MockFormData>({
  host: '',
  endpoint: '',
  method: 'GET',
  statusCode: 200,
  responseBody: '',
});
```

When editing, populate from mock:
```typescript
setFormData({
  host: mock.host,
  endpoint: path,
  method: mock.method as HttpMethod,
  statusCode: mock.statusCode,
  responseBody: '',
});
```

Add the form field between Method and Query Parameters:
```tsx
{/* Status Code */}
<div className="form-group">
  <label htmlFor="statusCode">Status Code</label>
  <input
    id="statusCode"
    type="number"
    className="form-input"
    min={100}
    max={599}
    value={formData.statusCode}
    onChange={(e) => handleChange('statusCode', e.target.value)}
    style={{ width: '120px' }}
  />
</div>
```

Add validation:
```typescript
const sc = Number(formData.statusCode);
if (isNaN(sc) || sc < 100 || sc > 599) {
  newErrors.statusCode = 'Status code must be between 100 and 599';
}
```

- [ ] **Step 2: Update HostEditModal.tsx**

Rename internal types:
- `ErrorRow` → `StatusRow`; field names unchanged (`code`, `percentage`)
- `ErrorRowError` → `StatusRowError`
- `UriState.errorRows` → `UriState.statusRows`
- `FormErrors.globalErrorRows` → `FormErrors.globalStatusRows`
- `FormErrors.globalErrorSum` → `FormErrors.globalStatusSum`
- `UriFormError.errorRows` → `UriFormError.statusRows`
- `UriFormError.errorSum` → `UriFormError.statusSum`

Update `validateErrorRows` → `validateStatusRows`:
- Change status code validation from `code < 400 || code > 599` to `code < 100 || code > 599`
- Update error message: "must be between 100 and 599"

Update all JSX labels:
- "Global Errors" → "Status Simulation"
- "Add Error" → "Add Status"
- "errors" in aria labels → "status"

Update `HostSaveData` construction to use `statuses:` instead of `errors:`.

- [ ] **Step 3: Update HostDetail.tsx**

```tsx
// Change:
const errorEntries = host.errors ? Object.entries(host.errors) : [];
// To:
const statusEntries = host.statuses ? Object.entries(host.statuses) : [];

// Change:
const hasConfig = host.latency || errorEntries.length > 0 || uriEntries.length > 0;
// To:
const hasConfig = host.latency || statusEntries.length > 0 || uriEntries.length > 0;

// Change section heading:
<h4>Global Errors</h4>
// To:
<h4>Status Simulation</h4>

// Update all errorEntries → statusEntries references

// In URI override display:
{uri.errors && Object.entries(uri.errors).map(...)
// To:
{uri.statuses && Object.entries(uri.statuses).map(...)
```

- [ ] **Step 4: Run frontend type check**

```bash
cd /Users/cseverino/workspaces/go-mock-server/web && npx tsc --noEmit 2>&1 | tail -30
```

Expected: no type errors.

- [ ] **Step 5: Commit**

```bash
cd /Users/cseverino/workspaces/go-mock-server
git add web/app/components/
git commit -m "feat: add status code field to MockEditModal and rename error simulation to status simulation in UI"
```

---

## Task 9: Update documentation and sample mocks

**Files:**
- Modify: `README.md`
- Modify: `roadmap.md`
- Modify: `sample-mocks/` — rename all mock files to add `.200` suffix
- Modify: any Swagger/OpenAPI spec files in `docs/`

- [ ] **Step 1: Check for Swagger files**

```bash
find /Users/cseverino/workspaces/go-mock-server/docs -name "*.yaml" -o -name "*.json" -o -name "*.yml" | sort
```

- [ ] **Step 2: Rename sample mock files**

```bash
cd /Users/cseverino/workspaces/go-mock-server/sample-mocks/example.host.com
for f in path1.get path1/root.get path2/post-uri.post; do
  mv "$f" "${f}.200"
done
```

Verify:
```bash
find /Users/cseverino/workspaces/go-mock-server/sample-mocks -type f | sort
```

- [ ] **Step 3: Update README.md**

Find and update all occurrences:
- "error simulation" → "status simulation"
- "error rate" → "status rate" or "status simulation percentage"
- `--disable-error` → `--disable-status-simulation`
- Mock file naming convention: update examples to show `.get.200` suffix
- Add **Breaking Changes** section:
  ```markdown
  ## Breaking Changes

  ### Mock file naming (v?.?.?)
  Mock files now require an explicit HTTP status code suffix.

  **Before:** `example.com/api/users.get`
  **After:** `example.com/api/users.get.200`

  To migrate, rename all existing mock files:
  ```bash
  find ./mocks -type f | grep -v '\.[0-9]\{3\}$' | while read f; do
    mv "$f" "${f}.200"
  done
  ```
  ```
- Update `/admin/config/hosts/{host}/errors` → `/admin/config/hosts/{host}/statuses` in API docs sections
- Document `_default.{status}` fallback file convention

- [ ] **Step 4: Update roadmap.md**

Replace any "error simulation" references with "status simulation".

- [ ] **Step 5: Update Swagger spec (if found)**

Update route descriptions, parameter names, and examples for:
- `/api/v1/config/hosts/{host}/statuses` (was `/errors`)
- `x-mock-status` header (new)
- `status_code` field in mock list response
- `statuses` field in host config

- [ ] **Step 6: Run full test suite**

```bash
cd /Users/cseverino/workspaces/go-mock-server && go test ./... 2>&1 | tail -20
```

Expected: all tests pass.

- [ ] **Step 7: Commit**

```bash
cd /Users/cseverino/workspaces/go-mock-server
git add README.md roadmap.md sample-mocks/ docs/
git commit -m "docs: rename error simulation to status simulation and document status-code mock file convention"
```

---

## Task 10: Full integration verification

- [ ] **Step 1: Build the binary**

```bash
cd /Users/cseverino/workspaces/go-mock-server && CGO_ENABLED=0 go build -o /tmp/mock-server ./cmd/mock-server/main.go 2>&1
```

Expected: clean build, no errors.

- [ ] **Step 2: Run the full test suite one final time**

```bash
cd /Users/cseverino/workspaces/go-mock-server && go test ./... -count=1 2>&1 | tail -20
```

Expected: all tests pass.

- [ ] **Step 3: Create a test mock and verify resolution**

```bash
mkdir -p /tmp/test-mocks/example.com/api
echo '{"status":"ok"}' > /tmp/test-mocks/example.com/api/users.get.200
echo '{"error":"internal"}' > /tmp/test-mocks/example.com/api/users.get.500
echo '{"error":"default"}' > /tmp/test-mocks/example.com/_default.500

/tmp/mock-server --mocks-directory /tmp/test-mocks --admin-port 0 &
SERVER_PID=$!
sleep 1

# Test 200 lookup
curl -s -H "Host: example.com" http://localhost:8080/api/users

# Test 500 via status simulation (configure via config file or test directly)
kill $SERVER_PID
```

Expected: 200 request returns `{"status":"ok"}`.

- [ ] **Step 4: Build frontend**

```bash
cd /Users/cseverino/workspaces/go-mock-server/web && npm ci && npm run build 2>&1 | tail -20
```

Expected: clean build.

- [ ] **Step 5: Final commit**

```bash
cd /Users/cseverino/workspaces/go-mock-server
git add -A
git commit -m "chore: final integration verification — status-code mocks feature complete"
```
