package middleware

import (
	"github.com/Caik/go-mock-server/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Uuid(ctx *gin.Context) {
	uuid := uuid.New()
	ctx.Set(util.UuidKey, uuid.String())

	ctx.Next()
}
