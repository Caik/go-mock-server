package admin

import "github.com/Caik/go-mock-server/internal/service/file"

type MockAddDeleteRequest struct {
	Host   string
	URI    string
	Method string
	Data   *[]byte
}

func AddUpdateMock(addRequest MockAddDeleteRequest) error {
	return file.SaveUpdateFile(addRequest.Host, addRequest.URI, addRequest.Method, addRequest.Data)
}

func DeleteMock(addRequest MockAddDeleteRequest) error {
	return file.DeleteFile(addRequest.Host, addRequest.URI, addRequest.Method)
}
