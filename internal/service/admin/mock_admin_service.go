package admin

import (
	"errors"

	"github.com/Caik/go-mock-server/internal/service/cache"
	"github.com/Caik/go-mock-server/internal/service/content"
	"github.com/Caik/go-mock-server/internal/service/mock"
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
		return errors.New("bad configuration found, content service is nil!")
	}

	err := contentService.SetContent(addRequest.Host, addRequest.URI, addRequest.Method, uuid, addRequest.Data)

	if err != nil {
		return err
	}

	cache.GetCacheService().Set(mock.GenerateCacheKey(mock.MockRequest{
		Host:   addRequest.Host,
		Method: addRequest.Method,
		URI:    addRequest.URI,
	}), addRequest.Data, uuid)

	return err
}

func DeleteMock(addRequest MockAddDeleteRequest, uuid string) error {
	contentService := content.GetContentService()

	if contentService == nil {
		return errors.New("bad configuration found, content service is nil!")
	}

	return contentService.DeleteContent(addRequest.Host, addRequest.URI, addRequest.Method, uuid)
}
