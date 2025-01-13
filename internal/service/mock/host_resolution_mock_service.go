package mock

import (
	"errors"
	"strings"
	"sync"

	"github.com/Caik/go-mock-server/internal/service/content"
	"github.com/Caik/go-mock-server/internal/util"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type hostResolutionMockService struct {
	next           mockService
	once           sync.Once
	contentService content.ContentService
	pathHosts      map[string]string
}

func (h *hostResolutionMockService) getMockResponse(mockRequest MockRequest) *MockResponse {
	mockRequest = h.evaluate(mockRequest)

	// calling next in the chain
	return h.nextOrNil(mockRequest)
}

func (h *hostResolutionMockService) evaluate(request MockRequest) MockRequest {
	if !util.IpAddressRegex.MatchString(request.Host) && util.HostRegex.MatchString(request.Host) {
		return request
	}

	host, exists := h.pathHosts[h.generateKey(request.Method, request.URI)]

	if !exists {
		return request
	}

	log.WithField("uuid", request.Uuid).
		WithField("host", request.Host).
		WithField("method", request.Method).
		WithField("uri", request.URI).
		WithField("new_host", host).
		Info("host resolved for request")

	return MockRequest{
		Host:   host,
		Method: request.Method,
		URI:    request.URI,
		Uuid:   request.Uuid,
	}
}

func (h *hostResolutionMockService) setNext(next mockService) {
	h.next = next
}

func (h *hostResolutionMockService) ensureInit(uuid string) error {
	if h.pathHosts != nil {
		return nil
	}

	h.once.Do(func() {
		h.pathHosts = make(map[string]string)
		data, err := h.contentService.ListContents(uuid)

		if err != nil {
			log.WithField("uuid", uuid).
				Errorf("error while trying to list contents: %v", err)

			return
		}

		for _, item := range *data {
			h.pathHosts[h.generateKey(item.Method, item.Uri)] = item.Host
		}

		channel := h.contentService.Subscribe("host_resolution_mock_service")

		go func() {
			log.WithField("uuid", uuid).
				Info("starting to listen for content changes")

			// listening to content change events
			for event := range channel {
				key := h.generateKey(event.Data.Method, event.Data.Uri)

				if event.Type == content.Removed {
					delete(h.pathHosts, key)
				} else {
					h.pathHosts[key] = event.Data.Host
				}
			}

			log.WithField("uuid", uuid).
				Info("stopping to listen for content changes")
		}()
	})

	return nil
}

func (h *hostResolutionMockService) generateKey(method, uri string) string {
	return strings.Join([]string{method, uri}, ":")
}

func (h *hostResolutionMockService) nextOrNil(mockRequest MockRequest) *MockResponse {
	if h.next == nil {
		return nil
	}

	return h.next.getMockResponse(mockRequest)
}

func newHostResolutionMockService(contentService content.ContentService) (*hostResolutionMockService, error) {
	uuid := uuid.NewString()

	if contentService == nil {
		log.WithField("uuid", uuid).
			Warn("bad configuration found, content service is nil!")

		return nil, errors.New("content service is nil")
	}

	service := hostResolutionMockService{
		contentService: contentService,
	}

	err := service.ensureInit(uuid)

	return &service, err
}
