package controller

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Caik/go-mock-server/internal/rest"
	"github.com/Caik/go-mock-server/internal/service/admin"
	"github.com/Caik/go-mock-server/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type AddDeleteMockRequest struct {
	Host   string `header:"x-mock-host" binding:"required"`
	Uri    string `header:"x-mock-uri" binding:"required"`
	Method string `header:"x-mock-method" binding:"required"`
}

type AdminMocksController struct {
	service *admin.MockAdminService
}

func (a *AdminMocksController) handleMocksList(c *gin.Context) {
	uuid := c.GetString(util.UuidKey)

	log.Info().
		Str("uuid", uuid).
		Msg("listing mocks")

	mocks, err := a.service.ListMocks(uuid)

	if err != nil {
		msg := fmt.Sprintf("error while listing mocks: %v", err)

		c.JSON(http.StatusInternalServerError, rest.Response{
			Status:  rest.Error,
			Message: msg,
		})

		log.Err(err).
			Stack().
			Str("uuid", uuid).
			Msg(msg)

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": rest.Success,
		"data":   mocks,
	})
}

func (a *AdminMocksController) handleMockContent(c *gin.Context) {
	uuid := c.GetString(util.UuidKey)
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, rest.Response{
			Status:  rest.Fail,
			Message: "missing required path param: id",
		})
		return
	}

	log.Info().
		Str("uuid", uuid).
		Str("id", id).
		Msg("getting mock content")

	content, err := a.service.GetMockContent(id, uuid)

	if err != nil {
		if errors.Is(err, admin.ErrInvalidMockID) {
			c.JSON(http.StatusBadRequest, rest.Response{
				Status:  rest.Fail,
				Message: err.Error(),
			})
			return
		}

		// ErrMockNotFound or other errors
		c.JSON(http.StatusNotFound, rest.Response{
			Status:  rest.Fail,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": rest.Success,
		"data": gin.H{
			"body": string(content),
		},
	})
}

func (a *AdminMocksController) handleMockAddUpdate(c *gin.Context) {
	req := AddDeleteMockRequest{}

	if err := c.ShouldBindHeader(&req); err != nil {
		c.JSON(http.StatusBadRequest, rest.Response{
			Status:  rest.Fail,
			Message: fmt.Sprintf("invalid request: %v", err),
		})

		return
	}

	if err := req.validate(); err != nil {
		c.JSON(http.StatusBadRequest, rest.Response{
			Status:  rest.Fail,
			Message: fmt.Sprintf("invalid request: %v", err),
		})

		return
	}

	data, err := io.ReadAll(c.Request.Body)

	if err != nil {
		c.JSON(http.StatusInternalServerError, rest.Response{
			Status:  rest.Error,
			Message: fmt.Sprintf("error while reading request body: %v", err),
		})

		return
	}

	if len(data) == 0 {
		c.JSON(http.StatusBadRequest, rest.Response{
			Status:  rest.Fail,
			Message: "invalid request: request body is empty",
		})

		return
	}

	uuid := c.GetString(util.UuidKey)

	log.Info().
		Str("uuid", uuid).
		Str("host", req.Host).
		Str("uri", req.Uri).
		Str("method", req.Method).
		Msg("adding/updating mock")

	err = a.service.AddUpdateMock(admin.MockAddDeleteRequest{
		Host:   req.Host,
		URI:    req.Uri,
		Method: req.Method,
		Data:   &data,
	}, uuid)

	if err != nil {
		c.JSON(http.StatusInternalServerError, rest.Response{
			Status:  rest.Error,
			Message: fmt.Sprintf("error while adding/updating mock: %v", err),
		})

		log.Err(err).
			Str("uuid", uuid).
			Str("host", req.Host).
			Str("uri", req.Uri).
			Str("method", req.Method).
			Msg("failed to add/update mock")

		return
	}

	c.JSON(http.StatusOK, rest.Response{
		Status:  rest.Success,
		Message: "mock added/updated successfully",
	})
}

func (a *AdminMocksController) handleMockCreate(c *gin.Context) {
	req := AddDeleteMockRequest{}

	if err := c.ShouldBindHeader(&req); err != nil {
		c.JSON(http.StatusBadRequest, rest.Response{
			Status:  rest.Fail,
			Message: fmt.Sprintf("invalid request: %v", err),
		})

		return
	}

	if err := req.validate(); err != nil {
		c.JSON(http.StatusBadRequest, rest.Response{
			Status:  rest.Fail,
			Message: fmt.Sprintf("invalid request: %v", err),
		})

		return
	}

	data, err := io.ReadAll(c.Request.Body)

	if err != nil {
		c.JSON(http.StatusInternalServerError, rest.Response{
			Status:  rest.Error,
			Message: fmt.Sprintf("error while reading request body: %v", err),
		})

		return
	}

	if len(data) == 0 {
		c.JSON(http.StatusBadRequest, rest.Response{
			Status:  rest.Fail,
			Message: "invalid request: request body is empty",
		})

		return
	}

	uuid := c.GetString(util.UuidKey)

	log.Info().
		Str("uuid", uuid).
		Str("host", req.Host).
		Str("uri", req.Uri).
		Str("method", req.Method).
		Msg("creating mock")

	err = a.service.AddUpdateMock(admin.MockAddDeleteRequest{
		Host:   req.Host,
		URI:    req.Uri,
		Method: req.Method,
		Data:   &data,
	}, uuid)

	if err != nil {
		c.JSON(http.StatusInternalServerError, rest.Response{
			Status:  rest.Error,
			Message: fmt.Sprintf("error while creating mock: %v", err),
		})

		log.Err(err).
			Str("uuid", uuid).
			Str("host", req.Host).
			Str("uri", req.Uri).
			Str("method", req.Method).
			Msg("failed to create mock")

		return
	}

	c.JSON(http.StatusCreated, rest.Response{
		Status:  rest.Success,
		Message: "mock created successfully",
	})
}

func (a *AdminMocksController) handleMockUpdate(c *gin.Context) {
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, rest.Response{
			Status:  rest.Fail,
			Message: "missing required path param: id",
		})

		return
	}

	req := AddDeleteMockRequest{}

	if err := c.ShouldBindHeader(&req); err != nil {
		c.JSON(http.StatusBadRequest, rest.Response{
			Status:  rest.Fail,
			Message: fmt.Sprintf("invalid request: %v", err),
		})

		return
	}

	if err := req.validate(); err != nil {
		c.JSON(http.StatusBadRequest, rest.Response{
			Status:  rest.Fail,
			Message: fmt.Sprintf("invalid request: %v", err),
		})

		return
	}

	data, err := io.ReadAll(c.Request.Body)

	if err != nil {
		c.JSON(http.StatusInternalServerError, rest.Response{
			Status:  rest.Error,
			Message: fmt.Sprintf("error while reading request body: %v", err),
		})

		return
	}

	if len(data) == 0 {
		c.JSON(http.StatusBadRequest, rest.Response{
			Status:  rest.Fail,
			Message: "invalid request: request body is empty",
		})

		return
	}

	uuid := c.GetString(util.UuidKey)

	log.Info().
		Str("uuid", uuid).
		Str("id", id).
		Str("host", req.Host).
		Str("uri", req.Uri).
		Str("method", req.Method).
		Msg("updating mock")

	// Delete the old mock first
	if err := a.service.DeleteMockByID(id, uuid); err != nil {
		if errors.Is(err, admin.ErrInvalidMockID) {
			c.JSON(http.StatusBadRequest, rest.Response{
				Status:  rest.Fail,
				Message: err.Error(),
			})

			return
		}
		// Log warning but continue - mock might not exist yet
		log.Warn().
			Err(err).
			Str("uuid", uuid).
			Str("id", id).
			Msg("failed to delete original mock during update")
	}

	// Create the new mock
	err = a.service.AddUpdateMock(admin.MockAddDeleteRequest{
		Host:   req.Host,
		URI:    req.Uri,
		Method: req.Method,
		Data:   &data,
	}, uuid)

	if err != nil {
		c.JSON(http.StatusInternalServerError, rest.Response{
			Status:  rest.Error,
			Message: fmt.Sprintf("error while updating mock: %v", err),
		})

		log.Err(err).
			Str("uuid", uuid).
			Str("id", id).
			Str("host", req.Host).
			Str("uri", req.Uri).
			Str("method", req.Method).
			Msg("failed to update mock")

		return
	}

	c.JSON(http.StatusOK, rest.Response{
		Status:  rest.Success,
		Message: "mock updated successfully",
	})
}

func (a *AdminMocksController) handleMockDelete(c *gin.Context) {
	addReq := AddDeleteMockRequest{}

	if err := c.ShouldBindHeader(&addReq); err != nil {
		c.JSON(http.StatusBadRequest, rest.Response{
			Status:  rest.Fail,
			Message: fmt.Sprintf("invalid request: %v", err),
		})

		return
	}

	if err := addReq.validate(); err != nil {
		c.JSON(http.StatusBadRequest, rest.Response{
			Status:  rest.Fail,
			Message: fmt.Sprintf("invalid request: %v", err),
		})

		return
	}

	uuid := c.GetString(util.UuidKey)

	log.Info().
		Str("uuid", uuid).
		Str("host", addReq.Host).
		Str("uri", addReq.Uri).
		Str("method", addReq.Method).
		Msg("deleting mock")

	err := a.service.DeleteMock(admin.MockAddDeleteRequest{
		Host:   addReq.Host,
		URI:    addReq.Uri,
		Method: addReq.Method,
	}, uuid)

	if err != nil {
		msg := fmt.Sprintf("error while deleting mock: %v", err)

		c.JSON(http.StatusInternalServerError, rest.Response{
			Status:  rest.Error,
			Message: msg,
		})

		log.Err(err).
			Stack().
			Str("uuid", uuid).
			Str("host", addReq.Host).
			Str("uri", addReq.Uri).
			Str("method", addReq.Method).
			Msg("")

		return
	}

	// if success, return back 200
	c.JSON(http.StatusOK, rest.Response{
		Status:  rest.Success,
		Message: "mock deleted with success",
	})
}

func (a *AddDeleteMockRequest) validate() error {
	a.Host = strings.ToLower(strings.TrimSpace(a.Host))

	if len(a.Host) == 0 {
		return errors.New("invalid host provided: it should not be empty")
	}

	if !util.HostRegex.MatchString(a.Host) {
		return errors.New("invalid host provided: it doesn't match a host pattern")
	}

	a.Uri = strings.TrimSpace(a.Uri)

	if len(a.Uri) == 0 {
		return errors.New("invalid uri provided: it should not be empty")
	}

	if !util.UriRegex.MatchString(a.Uri) {
		return errors.New("invalid uri provided: it doesn't match a host pattern")
	}

	a.Method = strings.ToUpper(strings.TrimSpace(a.Method))

	if len(a.Method) == 0 {
		return errors.New("invalid method provided: it should not be empty")
	}

	if !util.HttpMethodRegex.MatchString(a.Method) {
		return errors.New("invalid method provided: it should be a valid HTTP method")
	}

	return nil
}

func NewAdminMocksController(service *admin.MockAdminService) *AdminMocksController {
	return &AdminMocksController{
		service: service,
	}
}
