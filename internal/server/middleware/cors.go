package middleware

import (
	"github.com/gin-gonic/gin"
)

// Cors adds CORS headers to allow cross-origin requests from the web UI
func Cors(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH, HEAD")
	c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin, X-Mock-Host, X-Mock-Uri, X-Mock-Method")
	c.Header("Access-Control-Max-Age", "86400")

	// Handle preflight requests
	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(204)
		return
	}

	c.Next()
}
