package controller

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func InitUIRoutes(r *gin.Engine, uiDir string) {
	absUIDir, err := filepath.Abs(uiDir)

	if err != nil {
		panic("ui-dir: cannot resolve absolute path: " + err.Error())
	}

	r.GET("/ui/*path", func(c *gin.Context) {
		filePath := filepath.Join(absUIDir, filepath.Clean(c.Param("path")))

		// Reject any path that escapes the UI directory
		if !strings.HasPrefix(filePath, absUIDir+string(filepath.Separator)) {
			c.File(filepath.Join(absUIDir, "index.html"))
			return
		}

		info, err := os.Stat(filePath)

		if err == nil && !info.IsDir() {
			c.File(filePath)
			return
		}

		c.File(filepath.Join(absUIDir, "index.html"))
	})

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/ui/")
	})
}
