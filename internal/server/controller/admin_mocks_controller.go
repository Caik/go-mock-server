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

func (a *AdminMocksController) handleMockAddUpdate(c *gin.Context) {
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

	// reading the actual mock data
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
		Str("host", addReq.Host).
		Str("uri", addReq.Uri).
		Str("method", addReq.Method).
		Msg("adding/updating mock")

	err = a.service.AddUpdateMock(admin.MockAddDeleteRequest{
		Host:   addReq.Host,
		URI:    addReq.Uri,
		Method: addReq.Method,
		Data:   &data,
	}, uuid)

	if err != nil {
		msg := fmt.Sprintf("error while adding/updating mock: %v", err)

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
		Message: "mock updated with success",
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
