package middleware

import (
	"github.com/Caik/go-mock-server/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Uuid(ctx *gin.Context) {
	ctx.Set(util.UuidKey, uuid.NewString())

	ctx.Next()
}
