package admin

import (
	"github.com/Caik/go-mock-server/internal/service/content"
)

type MockAddDeleteRequest struct {
	Host   string
	URI    string
	Method string
	Data   *[]byte
}

type MockAdminService struct {
	contentService content.ContentService
}

func (m *MockAdminService) AddUpdateMock(addRequest MockAddDeleteRequest, uuid string) error {
	return m.contentService.SetContent(addRequest.Host, addRequest.URI, addRequest.Method, uuid, addRequest.Data)
}

func (m *MockAdminService) DeleteMock(addRequest MockAddDeleteRequest, uuid string) error {
	return m.contentService.DeleteContent(addRequest.Host, addRequest.URI, addRequest.Method, uuid)
}

func NewMockAdminService(contentService content.ContentService) *MockAdminService {
	return &MockAdminService{
		contentService: contentService,
	}
}
