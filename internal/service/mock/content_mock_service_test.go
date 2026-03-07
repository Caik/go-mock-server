package mock

import (
	"errors"
	"net/http"
	"testing"

	"github.com/Caik/go-mock-server/internal/service/content"
)

// errContentService returns errContentServiceNotFound to trigger the 500 path
type errContentService struct{}

func (e *errContentService) GetContent(host, uri, method, uuid string) (*content.ContentResult, error) {
	return nil, errContentServiceNotFound
}

func (e *errContentService) SetContent(host, uri, method, uuid string, data *[]byte) error {
	return nil
}

func (e *errContentService) DeleteContent(host, uri, method, uuid string) error {
	return nil
}

func (e *errContentService) ListContents(uuid string) (*[]content.ContentData, error) {
	return nil, nil
}

func (e *errContentService) Subscribe(subscriberId string, eventTypes ...content.ContentEventType) <-chan content.ContentEvent {
	return make(chan content.ContentEvent)
}

func (e *errContentService) Unsubscribe(subscriberId string) {}

func TestContentMockService_new500Response(t *testing.T) {
	t.Run("returns 500 when content service returns errContentServiceNotFound", func(t *testing.T) {
		svc := &contentMockService{contentService: &errContentService{}}

		req := MockRequest{
			Host:   "example.com",
			URI:    "/api/test",
			Method: "GET",
			Uuid:   "test-uuid",
		}

		resp := svc.getMockResponse(req)

		if resp == nil {
			t.Fatal("expected non-nil response")
		}

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", resp.StatusCode)
		}

		if resp.Metadata[MetadataMatched] != "false" {
			t.Errorf("expected Matched=false, got %q", resp.Metadata[MetadataMatched])
		}
	})
}

func TestContentMockService_new404Response(t *testing.T) {
	t.Run("returns 404 when content not found", func(t *testing.T) {
		svc := &contentMockService{
			contentService: &mockContentService{
				contents: make(map[string][]byte),
				events:   make(chan content.ContentEvent),
			},
		}

		req := MockRequest{
			Host:   "example.com",
			URI:    "/nonexistent",
			Method: "GET",
			Uuid:   "test-uuid",
		}

		resp := svc.getMockResponse(req)

		if resp == nil {
			t.Fatal("expected non-nil response")
		}

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", resp.StatusCode)
		}

		if resp.Metadata[MetadataMatched] != "false" {
			t.Errorf("expected Matched=false, got %q", resp.Metadata[MetadataMatched])
		}
	})
}

func TestContentMockService_getMockResponse_found(t *testing.T) {
	t.Run("returns 200 with metadata when content found", func(t *testing.T) {
		data := []byte(`{"key":"value"}`)
		svc := &contentMockService{
			contentService: &mockContentService{
				contents: map[string][]byte{
					"example.com:/api/test:GET": data,
				},
				events: make(chan content.ContentEvent),
			},
		}

		req := MockRequest{
			Host:   "example.com",
			URI:    "/api/test",
			Method: "GET",
			Uuid:   "test-uuid",
		}

		resp := svc.getMockResponse(req)

		if resp == nil {
			t.Fatal("expected non-nil response")
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		if resp.Metadata[MetadataMatched] != "true" {
			t.Errorf("expected Matched=true, got %q", resp.Metadata[MetadataMatched])
		}

		if resp.Metadata[MetadataSource] == "" {
			t.Error("expected non-empty Source")
		}

		if resp.Metadata[MetadataPath] == "" {
			t.Error("expected non-empty Path")
		}
	})
}

// Verify errContentServiceNotFound is distinct from regular errors
func TestErrContentServiceNotFound(t *testing.T) {
	err := errContentServiceNotFound

	if !errors.Is(err, errContentServiceNotFound) {
		t.Error("expected errors.Is to match errContentServiceNotFound")
	}

	if errors.Is(errors.New("other error"), errContentServiceNotFound) {
		t.Error("expected errors.Is to NOT match a different error")
	}
}
