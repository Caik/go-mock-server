package content

type ContentService interface {
	GetContent(host, uri, method, uuid string) (*ContentResult, error)
	SetContent(host, uri, method, uuid string, data *[]byte) error
	DeleteContent(host, uri, method, uuid string) error
	ListContents(uuid string) (*[]ContentData, error)
	Subscribe(subscriberId string, eventTypes ...ContentEventType) <-chan ContentEvent
	Unsubscribe(subscriberId string)
}

// ContentResult contains the result of a GetContent call
type ContentResult struct {
	Data   *[]byte
	Source string // e.g., "filesystem", "s3", "redis" - implementation-defined
	Path   string // filesystem path, S3 key, etc.
}

type ContentEvent struct {
	Type ContentEventType
	Data ContentData
}

type ContentData struct {
	Host   string
	Uri    string
	Method string
}

type ContentEventType int

const (
	Created ContentEventType = iota
	Updated
	Removed
)

func (c ContentEventType) String() string {
	switch c {
	case Created:
		return "CREATED"
	case Updated:
		return "UPDATED"
	case Removed:
		return "REMOVED"
	default:
		return ""
	}
}
