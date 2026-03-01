package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Caik/go-mock-server/internal/rest"
	"github.com/Caik/go-mock-server/internal/service/traffic"
	"github.com/Caik/go-mock-server/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// TrafficController handles traffic log streaming and queries
type TrafficController struct {
	trafficLogService *traffic.TrafficLogService
}

// handleTrafficStream handles SSE streaming of traffic logs
func (t *TrafficController) handleTrafficStream(c *gin.Context) {
	uuid := c.GetString(util.UuidKey)

	// Check if traffic logging is enabled
	if t.trafficLogService == nil {
		c.JSON(http.StatusServiceUnavailable, rest.Response{
			Status:  rest.Fail,
			Message: "traffic logging is disabled",
		})
		return
	}

	// Parse optional filters from query params
	filters, err := t.parseFilters(c)

	if err != nil {
		c.JSON(http.StatusBadRequest, rest.Response{
			Status:  rest.Fail,
			Message: fmt.Sprintf("invalid filter: %v", err),
		})
		return
	}

	log.Info().
		Str("uuid", uuid).
		Msg("starting traffic stream")

	// Set SSE headers
	t.addSSEHeaders(c)

	// Subscribe to live traffic
	subscriberID := fmt.Sprintf("sse-%s", uuid)
	ch := t.trafficLogService.Subscribe(subscriberID, filters)
	defer t.trafficLogService.Unsubscribe(subscriberID)

	// Send catch-up entries first
	catchUp := t.trafficLogService.GetFiltered(filters)

	for _, entry := range catchUp {
		if err := t.writeSSEEvent(c, entry, uuid); err != nil {
			return
		}
	}

	// Stream live entries
	clientGone := c.Request.Context().Done()

	for {
		select {
		case <-clientGone:
			log.Info().
				Str("uuid", uuid).
				Msg("client disconnected from traffic stream")

			return
		case entry, ok := <-ch:
			if !ok {
				log.Info().
					Str("uuid", uuid).
					Msg("traffic stream channel closed")

				return
			}
			if err := t.writeSSEEvent(c, entry, uuid); err != nil {
				return
			}
		}
	}
}

func (t *TrafficController) addSSEHeaders(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // Disable nginx buffering
}

// parseFilters extracts filter parameters from query string
func (t *TrafficController) parseFilters(c *gin.Context) (*traffic.TrafficFilters, error) {
	hostsParam := c.Query("hosts")
	statusParam := c.Query("status")
	matchedParam := c.Query("matched")

	// Return nil if no filters provided
	if hostsParam == "" && statusParam == "" && matchedParam == "" {
		return nil, nil
	}

	// Parse hosts (comma-separated)
	var hosts []string

	if hostsParam != "" {
		hosts = strings.Split(hostsParam, ",")
	}

	// Parse status codes (comma-separated)
	var statusCodes []int

	if statusParam != "" {
		codes := strings.Split(statusParam, ",")

		for _, code := range codes {
			statusCode, err := strconv.Atoi(strings.TrimSpace(code))

			if err != nil {
				return nil, fmt.Errorf("invalid status code: %s", code)
			}

			statusCodes = append(statusCodes, statusCode)
		}
	}

	// Parse matched filter
	var matched *bool

	if matchedParam != "" {
		matchedBool, err := strconv.ParseBool(matchedParam)

		if err != nil {
			return nil, fmt.Errorf("invalid matched value: %s", matchedParam)
		}

		matched = &matchedBool
	}

	// Build and validate filters
	filters := &traffic.TrafficFilters{
		Hosts:       hosts,
		StatusCodes: statusCodes,
		Matched:     matched,
	}

	if err := filters.Validate(); err != nil {
		return nil, err
	}

	return filters, nil
}

// writeSSEEvent marshals and writes a traffic entry as an SSE event.
// Returns nil on marshal errors (continue streaming), error on write errors (stop streaming).
func (t *TrafficController) writeSSEEvent(c *gin.Context, entry traffic.TrafficEntry, uuid string) error {
	data, err := json.Marshal(entry)

	if err != nil {
		log.Warn().
			Err(err).
			Stack().
			Str("uuid", uuid).
			Msg("failed to marshal entry")

		return nil
	}

	if _, err := fmt.Fprintf(c.Writer, "data: %s\n\n", data); err != nil {
		log.Warn().
			Err(err).
			Str("uuid", uuid).
			Msg("failed to write SSE event")

		return err
	}

	c.Writer.Flush()

	return nil
}

// NewTrafficController creates a new TrafficController
func NewTrafficController(trafficLogService *traffic.TrafficLogService) *TrafficController {
	return &TrafficController{
		trafficLogService: trafficLogService,
	}
}
