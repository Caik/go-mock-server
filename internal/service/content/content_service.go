package content

type ContentService interface {
	GetContent(host, uri, method, uuid string) (*[]byte, error)
	SetContent(host, uri, method, uuid string, data *[]byte) error
	DeleteContent(host, uri, method, uuid string) error
	ListContents(uuid string) (*[]ContentData, error)
	Subscribe(subscriberId string, eventTypes ...ContentEventType) <-chan ContentEvent
	Unsubscribe(subscriberId string)
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
