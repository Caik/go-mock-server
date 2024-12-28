package admin

import (
	"errors"

	"github.com/Caik/go-mock-server/internal/service/content"
)

type MockAddDeleteRequest struct {
	Host   string
	URI    string
	Method string
	Data   *[]byte
}

func AddUpdateMock(addRequest MockAddDeleteRequest, uuid string) error {
	contentService := content.GetContentService()

	if contentService == nil {
		return errors.New("bad configuration found, content service is nil")
	}

	err := contentService.SetContent(addRequest.Host, addRequest.URI, addRequest.Method, uuid, addRequest.Data)

	if err != nil {
		return err
	}

	return err
}

func DeleteMock(addRequest MockAddDeleteRequest, uuid string) error {
	contentService := content.GetContentService()

	if contentService == nil {
		return errors.New("bad configuration found, content service is nil")
	}

	return contentService.DeleteContent(addRequest.Host, addRequest.URI, addRequest.Method, uuid)
}
