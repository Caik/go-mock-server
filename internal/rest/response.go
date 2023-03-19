package rest

const (
	Success Status = "success"
	Fail    Status = "fail"
	Error   Status = "error"
)

type Response struct {
	Status  Status      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type Status string
