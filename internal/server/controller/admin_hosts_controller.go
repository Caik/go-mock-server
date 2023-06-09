package controller

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Caik/go-mock-server/internal/config"
	"github.com/Caik/go-mock-server/internal/rest"
	"github.com/Caik/go-mock-server/internal/service/admin"
	"github.com/Caik/go-mock-server/internal/util"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type AddDeleteGetHostRequest struct {
	Host          string                        `json:"host" binding:"required"`
	LatencyConfig *config.LatencyConfig         `json:"latency"`
	ErrorConfig   map[string]config.ErrorConfig `json:"errors"`
	UriConfig     map[string]config.UriConfig   `json:"uris"`
	errorCode     string
}

func initAdminHostsController(r *gin.RouterGroup) {
	r.GET("", handleHostsConfigList)
	r.POST("", handleHostConfigAddUpdate)

	r.GET("/:host", handleHostConfigRetrieve)
	r.DELETE("/:host", handleHostConfigDelete)

	r.POST("/:host/latencies", handleLatencyAddUpdate)
	r.DELETE("/:host/latencies", handleLatencyDelete)

	r.POST("/:host/errors", handleErrorsAddUpdate)
	r.DELETE("/:host/errors/:error", handleErrorDelete)

	r.POST("/:host/uris", handleUrisAddUpdate)
}

func handleHostsConfigList(c *gin.Context) {
	log.WithField("uuid", c.GetString(util.UuidKey)).
		Info("getting hosts config")

	config, err := admin.GetHostsConfig()

	if err != nil {
		msg := fmt.Sprintf("error while getting hosts config: %v", err)

		c.JSON(http.StatusInternalServerError, rest.Response{
			Status:  rest.Error,
			Message: msg,
		})

		log.WithField("uuid", c.GetString(util.UuidKey)).
			Error(msg)

		return
	}

	// if success, return back 200
	c.JSON(http.StatusOK, rest.Response{
		Status:  rest.Success,
		Message: "hosts config retrieved with success",
		Data:    config,
	})
}

func handleHostConfigAddUpdate(c *gin.Context) {
	addReq := AddDeleteGetHostRequest{}

	if err := c.ShouldBind(&addReq); err != nil {
		c.JSON(http.StatusBadRequest, rest.Response{
			Status:  rest.Fail,
			Message: fmt.Sprintf("invalid request: %v", err),
		})

		return
	}

	if err := addReq.validate(false, false, false); err != nil {
		c.JSON(http.StatusBadRequest, rest.Response{
			Status:  rest.Fail,
			Message: fmt.Sprintf("invalid request: %v", err),
		})

		return
	}

	log.WithField("uuid", c.GetString(util.UuidKey)).
		WithField("host", addReq.Host).
		Info("adding/updating host config")

	hostConfig, err := admin.AddUpdateHost(admin.HostAddDeleteRequest{
		Host:          addReq.Host,
		LatencyConfig: addReq.LatencyConfig,
		ErrorConfig:   addReq.ErrorConfig,
		UriConfig:     addReq.UriConfig,
	})

	if err != nil {
		msg := fmt.Sprintf("error while adding/updating host config: %v", err)

		c.JSON(http.StatusInternalServerError, rest.Response{
			Status:  rest.Error,
			Message: msg,
		})

		log.WithField("uuid", c.GetString(util.UuidKey)).
			WithField("host", addReq.Host).
			Error(msg)

		return
	}

	// if success, return back 200
	c.JSON(http.StatusOK, rest.Response{
		Status:  rest.Success,
		Message: "host config updated with success",
		Data:    hostConfig,
	})
}

func handleHostConfigRetrieve(c *gin.Context) {
	getReq := AddDeleteGetHostRequest{Host: c.Param("host")}

	if err := getReq.validate(false, false, false); err != nil {
		c.JSON(http.StatusBadRequest, rest.Response{
			Status:  rest.Fail,
			Message: fmt.Sprintf("invalid request: %v", err),
		})

		return
	}

	log.WithField("uuid", c.GetString(util.UuidKey)).
		WithField("host", getReq.Host).
		Info("getting host config")

	hostConfig, err := admin.GetHostConfig(getReq.Host)

	if err != nil {
		msg := fmt.Sprintf("error while getting host config: %v", err)

		c.JSON(http.StatusInternalServerError, rest.Response{
			Status:  rest.Error,
			Message: msg,
		})

		log.WithField("uuid", c.GetString(util.UuidKey)).
			Error(msg)

		return
	}

	// if config is null it means it doesn't exist
	if hostConfig == nil {
		c.JSON(http.StatusNotFound, rest.Response{
			Status:  rest.Fail,
			Message: "host config not found",
		})

		return
	}

	// if success, return back 200
	c.JSON(http.StatusOK, rest.Response{
		Status:  rest.Success,
		Message: "host config retrieved with success",
		Data:    hostConfig,
	})
}

func handleHostConfigDelete(c *gin.Context) {
	deleteReq := AddDeleteGetHostRequest{Host: c.Param("host")}

	if err := deleteReq.validate(false, false, false); err != nil {
		c.JSON(http.StatusBadRequest, rest.Response{
			Status:  rest.Fail,
			Message: fmt.Sprintf("invalid request: %v", err),
		})

		return
	}

	log.WithField("uuid", c.GetString(util.UuidKey)).
		WithField("host", deleteReq.Host).
		Info("deleting host config")

	err := admin.DeleteHost(deleteReq.Host)

	if err != nil {
		msg := fmt.Sprintf("error while deleting host config: %v", err)

		c.JSON(http.StatusInternalServerError, rest.Response{
			Status:  rest.Error,
			Message: msg,
		})

		log.WithField("uuid", c.GetString(util.UuidKey)).
			WithField("host", deleteReq.Host).
			Error(msg)

		return
	}

	// if success, return back 200
	c.JSON(http.StatusOK, rest.Response{
		Status:  rest.Success,
		Message: "host config deleted with success",
	})
}

func handleLatencyAddUpdate(c *gin.Context) {
	addLatencyReq := AddDeleteGetHostRequest{Host: c.Param("host")}

	if err := c.ShouldBind(&addLatencyReq); err != nil {
		c.JSON(http.StatusBadRequest, rest.Response{
			Status:  rest.Fail,
			Message: fmt.Sprintf("invalid request: %v", err),
		})

		return
	}

	if err := addLatencyReq.validate(true, false, false); err != nil {
		c.JSON(http.StatusBadRequest, rest.Response{
			Status:  rest.Fail,
			Message: fmt.Sprintf("invalid request: %v", err),
		})

		return
	}

	log.WithField("uuid", c.GetString(util.UuidKey)).
		WithField("host", addLatencyReq.Host).
		Info("adding/updating host latency config")

	hostConfig, err := admin.AddUpdateHostLatency(admin.HostAddDeleteRequest{
		Host:          addLatencyReq.Host,
		LatencyConfig: addLatencyReq.LatencyConfig,
	})

	if err != nil {
		msg := fmt.Sprintf("error while adding/updating host latency config: %v", err)

		c.JSON(http.StatusInternalServerError, rest.Response{
			Status:  rest.Error,
			Message: msg,
		})

		log.WithField("uuid", c.GetString(util.UuidKey)).
			WithField("host", addLatencyReq.Host).
			Error(msg)

		return
	}

	if hostConfig == nil {
		c.JSON(http.StatusNotFound, rest.Response{
			Status:  rest.Fail,
			Message: "host config not found",
		})

		return
	}

	// if success, return back 200
	c.JSON(http.StatusOK, rest.Response{
		Status:  rest.Success,
		Message: "host latency config updated with success",
		Data:    hostConfig,
	})
}

func handleLatencyDelete(c *gin.Context) {
	latencyDeleteReq := AddDeleteGetHostRequest{Host: c.Param("host")}

	if err := latencyDeleteReq.validate(false, false, false); err != nil {
		c.JSON(http.StatusBadRequest, rest.Response{
			Status:  rest.Fail,
			Message: fmt.Sprintf("invalid request: %v", err),
		})

		return
	}

	hostConfig, err := admin.DeleteHostLatency(latencyDeleteReq.Host)

	if err != nil {
		msg := fmt.Sprintf("error while deleting host latency config: %v", err)

		c.JSON(http.StatusInternalServerError, rest.Response{
			Status:  rest.Error,
			Message: msg,
		})

		log.WithField("uuid", c.GetString(util.UuidKey)).
			WithField("host", latencyDeleteReq.Host).
			Error(msg)

		return
	}

	if hostConfig == nil {
		c.JSON(http.StatusNotFound, rest.Response{
			Status:  rest.Fail,
			Message: "host config not found",
		})

		return
	}

	// if success, return back 200
	c.JSON(http.StatusOK, rest.Response{
		Status:  rest.Success,
		Message: "host latency config deleted with success",
		Data:    hostConfig,
	})
}

func handleErrorsAddUpdate(c *gin.Context) {
	addErrorsReq := AddDeleteGetHostRequest{Host: c.Param("host")}

	if err := c.ShouldBind(&addErrorsReq); err != nil {
		c.JSON(http.StatusBadRequest, rest.Response{
			Status:  rest.Fail,
			Message: fmt.Sprintf("invalid request: %v", err),
		})

		return
	}

	if err := addErrorsReq.validate(false, true, false); err != nil {
		c.JSON(http.StatusBadRequest, rest.Response{
			Status:  rest.Fail,
			Message: fmt.Sprintf("invalid request: %v", err),
		})

		return
	}

	log.WithField("uuid", c.GetString(util.UuidKey)).
		WithField("host", addErrorsReq.Host).
		Info("adding/updating host errors config")

	hostConfig, err := admin.AddUpdateHostErrors(admin.HostAddDeleteRequest{
		Host:        addErrorsReq.Host,
		ErrorConfig: addErrorsReq.ErrorConfig,
	})

	if err != nil {
		msg := fmt.Sprintf("error while adding/updating host errors config: %v", err)

		c.JSON(http.StatusInternalServerError, rest.Response{
			Status:  rest.Error,
			Message: msg,
		})

		log.WithField("uuid", c.GetString(util.UuidKey)).
			WithField("host", addErrorsReq.Host).
			Error(msg)

		return
	}

	if hostConfig == nil {
		c.JSON(http.StatusNotFound, rest.Response{
			Status:  rest.Fail,
			Message: "host config not found",
		})

		return
	}

	// if success, return back 200
	c.JSON(http.StatusOK, rest.Response{
		Status:  rest.Success,
		Message: "host errors config updated with success",
		Data:    hostConfig,
	})
}

func handleErrorDelete(c *gin.Context) {
	errorDeleteReq := AddDeleteGetHostRequest{Host: c.Param("host"), errorCode: c.Param("error")}

	if err := errorDeleteReq.validate(false, false, true); err != nil {
		c.JSON(http.StatusBadRequest, rest.Response{
			Status:  rest.Fail,
			Message: fmt.Sprintf("invalid request: %v", err),
		})

		return
	}

	hostConfig, err := admin.DeleteHostError(errorDeleteReq.Host, errorDeleteReq.errorCode)

	if err != nil {
		msg := fmt.Sprintf("error while deleting host latency config: %v", err)

		c.JSON(http.StatusInternalServerError, rest.Response{
			Status:  rest.Error,
			Message: msg,
		})

		log.WithField("uuid", c.GetString(util.UuidKey)).
			WithField("host", errorDeleteReq.Host).
			Error(msg)

		return
	}

	if hostConfig == nil {
		c.JSON(http.StatusNotFound, rest.Response{
			Status:  rest.Fail,
			Message: "host config not found",
		})

		return
	}

	// if success, return back 200
	c.JSON(http.StatusOK, rest.Response{
		Status:  rest.Success,
		Message: "host error config deleted with success",
		Data:    hostConfig,
	})
}

func handleUrisAddUpdate(c *gin.Context) {
	addErrorsReq := AddDeleteGetHostRequest{Host: c.Param("host")}

	if err := c.ShouldBind(&addErrorsReq); err != nil {
		c.JSON(http.StatusBadRequest, rest.Response{
			Status:  rest.Fail,
			Message: fmt.Sprintf("invalid request: %v", err),
		})

		return
	}

	if err := addErrorsReq.validate(false, false, false); err != nil {
		c.JSON(http.StatusBadRequest, rest.Response{
			Status:  rest.Fail,
			Message: fmt.Sprintf("invalid request: %v", err),
		})

		return
	}

	log.WithField("uuid", c.GetString(util.UuidKey)).
		WithField("host", addErrorsReq.Host).
		Info("adding/updating host uris config")

	hostConfig, err := admin.AddUpdateHostUris(admin.HostAddDeleteRequest{
		Host:      addErrorsReq.Host,
		UriConfig: addErrorsReq.UriConfig,
	})

	if err != nil {
		msg := fmt.Sprintf("error while adding/updating host uris config: %v", err)

		c.JSON(http.StatusInternalServerError, rest.Response{
			Status:  rest.Error,
			Message: msg,
		})

		log.WithField("uuid", c.GetString(util.UuidKey)).
			WithField("host", addErrorsReq.Host).
			Error(msg)

		return
	}

	if hostConfig == nil {
		c.JSON(http.StatusNotFound, rest.Response{
			Status:  rest.Fail,
			Message: "host config not found",
		})

		return
	}

	// if success, return back 200
	c.JSON(http.StatusOK, rest.Response{
		Status:  rest.Success,
		Message: "host uris config updated with success",
		Data:    hostConfig,
	})
}

func (a *AddDeleteGetHostRequest) validate(needsLatency, needsErrors, needsErrorCode bool) error {
	a.Host = strings.ToLower(strings.TrimSpace(a.Host))

	if len(a.Host) == 0 {
		return errors.New("invalid host provided: it should not be empty")
	}

	if !util.HostRegex.MatchString(a.Host) {
		return errors.New("invalid host provided: it doesn't match a host pattern")
	}

	if needsLatency && a.LatencyConfig == nil {
		return errors.New("invalid latency provided: it should not be empty")
	}

	if needsErrors && len(a.ErrorConfig) == 0 {
		return errors.New("invalid errors provided: it should not be empty")
	}

	if needsErrorCode && len(a.errorCode) == 0 {
		return errors.New("invalid error provided: it should not be empty")
	}

	return nil
}
