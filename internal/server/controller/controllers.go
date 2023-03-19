package controller

import "github.com/gin-gonic/gin"

func Init(r *gin.Engine) {
	initAdminMocksController(r.Group("/admin/mocks"))
	initAdminHostsController(r.Group("/admin/config/hosts"))

	r.NoRoute(handleMockRequest)
}
