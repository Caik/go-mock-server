package mock

import (
	"testing"

	"github.com/gin-gonic/gin"
)

var (
	ctms = newContentTypeMockService(MockServiceParams{
		defaultContentType: gin.MIMEPlain,
	})
)

func TestSetAppropriateContentType_JSON(t *testing.T) {
	mime := ctms.setAppropriateContentType("application/json")
	if mime != gin.MIMEJSON {
		t.Errorf(
			"expected %s, got %s",
			gin.MIMEJSON,
			mime,
		)
	}
}

func TestSetAppropriateContentType_PLAIN_FROM_ANY(t *testing.T) {
	mime := ctms.setAppropriateContentType("*/*")
	if mime != gin.MIMEPlain {
		t.Errorf(
			"expected %s, got %s",
			gin.MIMEJSON,
			mime,
		)
	}
}

func TestSetAppropriateContentType_PLAIN_FROM_EMPTY(t *testing.T) {
	mime := ctms.setAppropriateContentType("")
	if mime != gin.MIMEPlain {
		t.Errorf(
			"expected %s, got %s",
			gin.MIMEJSON,
			mime,
		)
	}
}

func TestSetAppropriateContentType_JSON_HTML(t *testing.T) {
	mime := ctms.setAppropriateContentType("application/json, text/html, */*; q=0.1")
	if mime != gin.MIMEJSON {
		t.Errorf(
			"expected %s, got %s",
			gin.MIMEJSON,
			mime,
		)
	}
}
