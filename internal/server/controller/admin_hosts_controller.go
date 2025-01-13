package controller

import (
	"errors"
	"fmt"
	"github.com/Caik/go-mock-server/internal/config"
	"net/http"
	"strings"

	"github.com/Caik/go-mock-server/internal/rest"
	"github.com/Caik/go-mock-server/internal/service/admin"
	"github.com/Caik/go-mock-server/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type AddDeleteGetHostRequest struct {
	Host          string                        `json:"host" binding:"required"`
	LatencyConfig *config.LatencyConfig         `json:"latency"`
	ErrorConfig   map[string]config.ErrorConfig `json:"errors"`
	UriConfig     map[string]config.UriConfig   `json:"uris"`
	errorCode     string
}

type AdminHostsController struct {
	hostsConfig *config.HostsConfig
	service     *admin.HostsConfigAdminService
}

func (a *AdminHostsController) handleHostsConfigList(c *gin.Context) {
	log.Info().
		Str("uuid", c.GetString(util.UuidKey)).
		Msg("getting hosts config")

	c.JSON(http.StatusOK, rest.Response{
		Status:  rest.Success,
		Message: "hosts config retrieved with success",
		Data:    a.hostsConfig,
	})
}

func (a *AdminHostsController) handleHostConfigAddUpdate(c *gin.Context) {
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

	log.Info().
		Str("uuid", c.GetString(util.UuidKey)).
		Str("host", addReq.Host).
		Msg("adding/updating host config")

	hostConfig, err := a.service.AddUpdateHost(admin.HostAddDeleteRequest{
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

		log.Err(err).
			Stack().
			Str("uuid", c.GetString(util.UuidKey)).
			Str("host", addReq.Host).
			Msg("")

		return
	}

	// if success, return back 200
	c.JSON(http.StatusOK, rest.Response{
		Status:  rest.Success,
		Message: "host config updated with success",
		Data:    hostConfig,
	})
}

func (a *AdminHostsController) handleHostConfigRetrieve(c *gin.Context) {
	getReq := AddDeleteGetHostRequest{Host: c.Param("host")}

	if err := getReq.validate(false, false, false); err != nil {
		c.JSON(http.StatusBadRequest, rest.Response{
			Status:  rest.Fail,
			Message: fmt.Sprintf("invalid request: %v", err),
		})

		return
	}

	log.Info().
		Str("uuid", c.GetString(util.UuidKey)).
		Str("host", getReq.Host).
		Msg("getting host config")

	hostConfig := a.service.GetHostConfig(getReq.Host)

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

func (a *AdminHostsController) handleHostConfigDelete(c *gin.Context) {
	deleteReq := AddDeleteGetHostRequest{Host: c.Param("host")}

	if err := deleteReq.validate(false, false, false); err != nil {
		c.JSON(http.StatusBadRequest, rest.Response{
			Status:  rest.Fail,
			Message: fmt.Sprintf("invalid request: %v", err),
		})

		return
	}

	log.Info().
		Str("uuid", c.GetString(util.UuidKey)).
		Str("host", deleteReq.Host).
		Msg("deleting host config")

	a.service.DeleteHost(deleteReq.Host)

	c.JSON(http.StatusOK, rest.Response{
		Status:  rest.Success,
		Message: "host config deleted with success",
	})
}

func (a *AdminHostsController) handleLatencyAddUpdate(c *gin.Context) {
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

	log.Info().
		Str("uuid", c.GetString(util.UuidKey)).
		Str("host", addLatencyReq.Host).
		Msg("adding/updating host latency config")

	hostConfig, err := a.service.AddUpdateHostLatency(admin.HostAddDeleteRequest{
		Host:          addLatencyReq.Host,
		LatencyConfig: addLatencyReq.LatencyConfig,
	})

	if err != nil {
		msg := fmt.Sprintf("error while adding/updating host latency config: %v", err)

		c.JSON(http.StatusInternalServerError, rest.Response{
			Status:  rest.Error,
			Message: msg,
		})

		log.Err(err).
			Stack().
			Str("uuid", c.GetString(util.UuidKey)).
			Str("host", addLatencyReq.Host).
			Msg("")

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

func (a *AdminHostsController) handleLatencyDelete(c *gin.Context) {
	latencyDeleteReq := AddDeleteGetHostRequest{Host: c.Param("host")}

	if err := latencyDeleteReq.validate(false, false, false); err != nil {
		c.JSON(http.StatusBadRequest, rest.Response{
			Status:  rest.Fail,
			Message: fmt.Sprintf("invalid request: %v", err),
		})

		return
	}

	hostConfig, err := a.service.DeleteHostLatency(latencyDeleteReq.Host)

	if err != nil {
		msg := fmt.Sprintf("error while deleting host latency config: %v", err)

		c.JSON(http.StatusInternalServerError, rest.Response{
			Status:  rest.Error,
			Message: msg,
		})

		log.Err(err).
			Stack().
			Str("uuid", c.GetString(util.UuidKey)).
			Str("host", latencyDeleteReq.Host).
			Msg("")

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

func (a *AdminHostsController) handleErrorsAddUpdate(c *gin.Context) {
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

	log.Info().
		Str("uuid", c.GetString(util.UuidKey)).
		Str("host", addErrorsReq.Host).
		Msg("adding/updating host errors config")

	hostConfig, err := a.service.AddUpdateHostErrors(admin.HostAddDeleteRequest{
		Host:        addErrorsReq.Host,
		ErrorConfig: addErrorsReq.ErrorConfig,
	})

	if err != nil {
		msg := fmt.Sprintf("error while adding/updating host errors config: %v", err)

		c.JSON(http.StatusInternalServerError, rest.Response{
			Status:  rest.Error,
			Message: msg,
		})

		log.Err(err).
			Stack().
			Str("uuid", c.GetString(util.UuidKey)).
			Str("host", addErrorsReq.Host).
			Msg("")

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

func (a *AdminHostsController) handleErrorDelete(c *gin.Context) {
	errorDeleteReq := AddDeleteGetHostRequest{Host: c.Param("host"), errorCode: c.Param("error")}

	if err := errorDeleteReq.validate(false, false, true); err != nil {
		c.JSON(http.StatusBadRequest, rest.Response{
			Status:  rest.Fail,
			Message: fmt.Sprintf("invalid request: %v", err),
		})

		return
	}

	hostConfig, err := a.service.DeleteHostError(errorDeleteReq.Host, errorDeleteReq.errorCode)

	if err != nil {
		msg := fmt.Sprintf("error while deleting host latency config: %v", err)

		c.JSON(http.StatusInternalServerError, rest.Response{
			Status:  rest.Error,
			Message: msg,
		})

		log.Err(err).
			Stack().
			Str("uuid", c.GetString(util.UuidKey)).
			Str("host", errorDeleteReq.Host).
			Msg("")

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

func (a *AdminHostsController) handleUrisAddUpdate(c *gin.Context) {
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

	log.Info().
		Str("uuid", c.GetString(util.UuidKey)).
		Str("host", addErrorsReq.Host).
		Msg("adding/updating host uris config")

	hostConfig, err := a.service.AddUpdateHostUris(admin.HostAddDeleteRequest{
		Host:      addErrorsReq.Host,
		UriConfig: addErrorsReq.UriConfig,
	})

	if err != nil {
		msg := fmt.Sprintf("error while adding/updating host uris config: %v", err)

		c.JSON(http.StatusInternalServerError, rest.Response{
			Status:  rest.Error,
			Message: msg,
		})

		log.Err(err).
			Stack().
			Str("uuid", c.GetString(util.UuidKey)).
			Str("host", addErrorsReq.Host).
			Msg("")

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

func NewAdminHostsController(hostsConfig *config.HostsConfig, service *admin.HostsConfigAdminService) *AdminHostsController {
	return &AdminHostsController{
		hostsConfig: hostsConfig,
		service:     service,
	}
}
