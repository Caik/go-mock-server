package traffic

import (
	"fmt"
	"strings"

	"github.com/Caik/go-mock-server/internal/config"
	"github.com/Caik/go-mock-server/internal/util"
	"github.com/rs/zerolog/log"
)

const metadataMatched = "Matched"

// TrafficLogService manages traffic logging with an in-memory ring buffer
// and broadcasts new entries to subscribers for real-time streaming.
type TrafficLogService struct {
	ringBuffer  *util.RingBuffer[TrafficEntry]
	broadcaster *util.Broadcaster[TrafficEntry]
}

// NewTrafficLogService creates a new TrafficLogService from AppArguments.
// Returns nil if traffic logging is disabled (bufferSize <= 0).
func NewTrafficLogService(args *config.AppArguments) *TrafficLogService {
	if args.TrafficLogBufferSize <= 0 {
		return nil
	}

	ringBuffer, err := util.NewRingBuffer[TrafficEntry](args.TrafficLogBufferSize)

	if err != nil {
		// This should never happen since we check bufferSize > 0 above
		log.Warn().
			Err(err).
			Stack().
			Int("buffer_size", args.TrafficLogBufferSize).
			Msg("unexpected error creating ring buffer for traffic log service, disabling traffic logging")

		return nil
	}

	return &TrafficLogService{
		ringBuffer:  ringBuffer,
		broadcaster: &util.Broadcaster[TrafficEntry]{},
	}
}

// Capture adds a traffic entry to the log and broadcasts it to subscribers.
func (t *TrafficLogService) Capture(entry TrafficEntry) {
	if t == nil {
		return
	}

	t.ringBuffer.Add(entry)
	t.broadcaster.PublishAsync(entry, entry.UUID)
}

// GetAll returns all entries in the buffer, ordered from oldest to newest.
func (t *TrafficLogService) GetAll() []TrafficEntry {
	if t == nil {
		return []TrafficEntry{}
	}

	return t.ringBuffer.GetAll()
}

// GetRecent returns the n most recent entries, ordered from oldest to newest.
func (t *TrafficLogService) GetRecent(n int) []TrafficEntry {
	if t == nil {
		return []TrafficEntry{}
	}

	return t.ringBuffer.GetRecent(n)
}

// GetFiltered returns entries matching the provided filters.
// If filters is nil or empty, all entries are returned.
func (t *TrafficLogService) GetFiltered(filters *TrafficFilters) []TrafficEntry {
	if t == nil {
		return []TrafficEntry{}
	}

	all := t.ringBuffer.GetAll()

	if filters == nil || filters.IsEmpty() {
		return all
	}

	var filtered []TrafficEntry

	for _, entry := range all {
		if filters.Matches(entry) {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

// Subscribe returns a channel that receives new traffic entries.
// If filters is nil or empty, all entries are received.
// Otherwise, only entries matching the filter are received.
// Returns nil if the service is nil.
func (t *TrafficLogService) Subscribe(subscriberID string, filters *TrafficFilters) <-chan TrafficEntry {
	if t == nil {
		return nil
	}

	acceptFn := func(event TrafficEntry) bool {
		if filters == nil || filters.IsEmpty() {
			return true
		}

		return filters.Matches(event)
	}

	return t.broadcaster.Subscribe(subscriberID, acceptFn)
}

// Unsubscribe removes a subscriber.
func (t *TrafficLogService) Unsubscribe(subscriberID string) {
	if t == nil {
		return
	}

	t.broadcaster.Unsubscribe(subscriberID)
}

// Clear removes all entries from the buffer.
func (t *TrafficLogService) Clear() {
	if t == nil {
		return
	}

	t.ringBuffer.Clear()
}

// Size returns the current number of entries in the buffer.
func (t *TrafficLogService) Size() int {
	if t == nil {
		return 0
	}

	return t.ringBuffer.Size()
}

// TrafficFilters contains optional filters for querying traffic entries.
type TrafficFilters struct {
	Hosts       []string // Match any of these hosts (case-insensitive)
	StatusCodes []int    // Match any of these status codes
	Matched     *bool    // Match entries by matched status
}

// Validate checks if the filter values are valid.
// Returns an error describing the first invalid value found.
func (f TrafficFilters) Validate() error {
	for _, code := range f.StatusCodes {
		if code < 100 || code > 599 {
			return fmt.Errorf("invalid status code %d: must be between 100 and 599", code)
		}
	}

	for _, host := range f.Hosts {
		host = strings.TrimSpace(host)

		if host == "" {
			return fmt.Errorf("invalid host: empty or whitespace-only host not allowed")
		}

		if !util.HostRegex.MatchString(host) && !util.IpAddressRegex.MatchString(host) {
			return fmt.Errorf("invalid host %q: must be a valid hostname or IP address", host)
		}
	}

	return nil
}

// IsEmpty returns true if no filters are set.
func (f TrafficFilters) IsEmpty() bool {
	return len(f.Hosts) == 0 && len(f.StatusCodes) == 0 && f.Matched == nil
}

// Matches returns true if the entry matches all non-empty filters.
// For array filters (Hosts, StatusCodes), the entry must match ANY value in the array.
func (f TrafficFilters) Matches(entry TrafficEntry) bool {
	if len(f.Hosts) > 0 && !f.matchesHost(entry.Request.Host) {
		return false
	}

	if len(f.StatusCodes) > 0 && !f.matchesStatusCode(entry.Response.StatusCode) {
		return false
	}

	if f.Matched != nil && (entry.Metadata[metadataMatched] == "true") != *f.Matched {
		return false
	}

	return true
}

// matchesHost returns true if the given host matches any host in the filter.
func (f TrafficFilters) matchesHost(host string) bool {
	for _, h := range f.Hosts {
		if strings.EqualFold(host, h) {
			return true
		}
	}
	return false
}

// matchesStatusCode returns true if the given status code matches any code in the filter.
func (f TrafficFilters) matchesStatusCode(code int) bool {
	for _, c := range f.StatusCodes {
		if code == c {
			return true
		}
	}
	return false
}
