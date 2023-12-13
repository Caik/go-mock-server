package middleware

import (
	"fmt"
	"time"

	"github.com/Caik/go-mock-server/internal/util"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func Logger(ctx *gin.Context) {
	log.WithField("uuid", ctx.GetString(util.UuidKey)).
		WithField("host", ctx.Request.Host).
		WithField("uri", ctx.Request.RequestURI).
		WithField("method", ctx.Request.Method).
		Info("request received")

	start := time.Now()

	ctx.Next()

	end := time.Now()

	log.WithField("uuid", ctx.GetString(util.UuidKey)).
		WithField("host", ctx.Request.Host).
		WithField("uri", ctx.Request.RequestURI).
		WithField("method", ctx.Request.Method).
		WithField("status_code", ctx.Writer.Status()).
		WithField("latency", fmt.Sprintf("%v", end.Sub(start))).
		Info("request finished")
}
