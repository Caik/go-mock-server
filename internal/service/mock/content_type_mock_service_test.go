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
	actual := ctms.setAppropriateContentType("application/json")
	expected := gin.MIMEJSON
	if actual != expected {
		t.Errorf(
			"expected %s, got %s",
			expected,
			actual,
		)
	}
}

func TestSetAppropriateContentType_PLAIN_FROM_ANY(t *testing.T) {
	actual := ctms.setAppropriateContentType("*/*")
	expected := gin.MIMEPlain
	if actual != expected {
		t.Errorf(
			"expected %s, got %s",
			expected,
			actual,
		)
	}
}

func TestSetAppropriateContentType_PLAIN_FROM_EMPTY(t *testing.T) {
	actual := ctms.setAppropriateContentType("")
	expected := gin.MIMEPlain
	if actual != expected {
		t.Errorf(
			"expected %s, got %s",
			expected,
			actual,
		)
	}
}

func TestSetAppropriateContentType_JSON_HTML(t *testing.T) {
	actual := ctms.setAppropriateContentType("application/json, text/html, */*; q=0.1")
	expected := gin.MIMEJSON
	if actual != expected {
		t.Errorf(
			"expected %s, got %s",
			expected,
			actual,
		)
	}
}
