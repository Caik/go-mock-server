package middleware

import (
	"fmt"
	"time"

	"github.com/Caik/go-mock-server/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func Logger(ctx *gin.Context) {
	log.Info().
		Str("uuid", ctx.GetString(util.UuidKey)).
		Str("host", ctx.Request.Host).
		Str("uri", ctx.Request.RequestURI).
		Str("method", ctx.Request.Method).
		Msg("request received")

	start := time.Now()

	ctx.Next()

	end := time.Now()

	log.Info().
		Str("uuid", ctx.GetString(util.UuidKey)).
		Str("host", ctx.Request.Host).
		Str("uri", ctx.Request.RequestURI).
		Str("method", ctx.Request.Method).
		Int("status_code", ctx.Writer.Status()).
		Str("latency", fmt.Sprintf("%v", end.Sub(start))).
		Msg("request finished")
}
