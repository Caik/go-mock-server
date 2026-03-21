package admin

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Caik/go-mock-server/internal/service/content"
)

const mockIDSeparator = "|"

// Error types for GetMockContent
var (
	ErrInvalidMockID = errors.New("invalid mock ID")
	ErrMockNotFound  = errors.New("mock not found")
)

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

type MockAdminService struct {
	contentService content.ContentService
}

func (m *MockAdminService) AddUpdateMock(addRequest MockAddDeleteRequest, uuid string) error {
	return m.contentService.SetContent(addRequest.Host, addRequest.URI, addRequest.Method, uuid, addRequest.StatusCode, addRequest.Data)
}

func (m *MockAdminService) DeleteMock(addRequest MockAddDeleteRequest, uuid string) error {
	return m.contentService.DeleteContent(addRequest.Host, addRequest.URI, addRequest.Method, uuid, addRequest.StatusCode)
}

func (m *MockAdminService) DeleteMockByID(id, uuid string) error {
	host, uri, method, statusCode, err := decodeMockID(id)

	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidMockID, err)
	}

	return m.contentService.DeleteContent(host, uri, method, uuid, statusCode)
}

func (m *MockAdminService) GetMockContent(id, uuid string) ([]byte, error) {
	// Decode the ID to get host, uri, method, statusCode
	host, uri, method, statusCode, err := decodeMockID(id)

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidMockID, err)
	}

	// Get the content
	result, err := m.contentService.GetContent(host, uri, method, uuid, statusCode)

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrMockNotFound, err)
	}

	if result == nil || result.Data == nil {
		return nil, ErrMockNotFound
	}

	return *result.Data, nil
}

func (m *MockAdminService) ListMocks(uuid string) ([]MockListItem, error) {
	contents, err := m.contentService.ListContents(uuid)

	if err != nil {
		return nil, err
	}

	if contents == nil {
		return []MockListItem{}, nil
	}

	mocks := make([]MockListItem, len(*contents))

	for i, c := range *contents {
		mocks[i] = MockListItem{
			ID:         generateMockID(c.Host, c.Uri, c.Method, c.StatusCode),
			Host:       c.Host,
			URI:        c.Uri,
			Method:     c.Method,
			StatusCode: c.StatusCode,
		}
	}

	return mocks, nil
}

// generateMockID creates a unique identifier for a mock based on its host, uri, method, and statusCode.
func generateMockID(host, uri, method string, statusCode int) string {
	data := fmt.Sprintf("%s%s%s%s%s%s%d",
		host, mockIDSeparator, uri, mockIDSeparator, method, mockIDSeparator, statusCode)
	return base64.URLEncoding.EncodeToString([]byte(data))
}

// decodeMockID decodes a mock ID back to host, uri, method, and statusCode.
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

func NewMockAdminService(contentService content.ContentService) *MockAdminService {
	return &MockAdminService{
		contentService: contentService,
	}
}
