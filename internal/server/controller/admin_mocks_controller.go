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
	log "github.com/sirupsen/logrus"
)

type AddDeleteMockRequest struct {
	Host   string `header:"x-mock-host" binding:"required"`
	Uri    string `header:"x-mock-uri" binding:"required"`
	Method string `header:"x-mock-method" binding:"required"`
}

func initAdminMocksController(r *gin.RouterGroup) {
	r.POST("", handleMockAddUpdate)
	r.DELETE("", handleMockDelete)
}

func handleMockAddUpdate(c *gin.Context) {
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

	log.WithField("uuid", c.GetString(util.UuidKey)).
		WithField("host", addReq.Host).
		WithField("uri", addReq.Uri).
		WithField("method", addReq.Method).
		Info("adding/updating mock")

	err = admin.AddUpdateMock(admin.MockAddDeleteRequest{
		Host:   addReq.Host,
		URI:    addReq.Uri,
		Method: addReq.Method,
		Data:   &data,
	})

	if err != nil {
		msg := fmt.Sprintf("error while adding/updating mock: %v", err)

		c.JSON(http.StatusInternalServerError, rest.Response{
			Status:  rest.Error,
			Message: msg,
		})

		log.WithField("uuid", c.GetString(util.UuidKey)).
			WithField("host", addReq.Host).
			WithField("uri", addReq.Uri).
			WithField("method", addReq.Method).
			Error(msg)

		return
	}

	// if success, return back 200
	c.JSON(http.StatusOK, rest.Response{
		Status:  rest.Success,
		Message: "mock updated with success",
	})
}

func handleMockDelete(c *gin.Context) {
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

	log.WithField("uuid", c.GetString(util.UuidKey)).
		WithField("host", addReq.Host).
		WithField("uri", addReq.Uri).
		WithField("method", addReq.Method).
		Info("deleting mock")

	err := admin.DeleteMock(admin.MockAddDeleteRequest{
		Host:   addReq.Host,
		URI:    addReq.Uri,
		Method: addReq.Method,
	})

	if err != nil {
		msg := fmt.Sprintf("error while deleting mock: %v", err)

		c.JSON(http.StatusInternalServerError, rest.Response{
			Status:  rest.Error,
			Message: msg,
		})

		log.WithField("uuid", c.GetString(util.UuidKey)).
			WithField("host", addReq.Host).
			WithField("uri", addReq.Uri).
			WithField("method", addReq.Method).
			Error(msg)

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
