package controller

import (
	"github.com/gin-gonic/gin"
)

func InitRoutes(r *gin.Engine, adminMocksController *AdminMocksController, adminHostsController *AdminHostsController, mocksController *MocksController) {
	initAdminMocksController(r.Group("/admin/mocks"), adminMocksController)
	initAdminHostsController(r.Group("/admin/config/hosts"), adminHostsController)

	r.NoRoute(mocksController.handleMockRequest)
}

func initAdminMocksController(r *gin.RouterGroup, controller *AdminMocksController) {
	r.POST("", controller.handleMockAddUpdate)
	r.DELETE("", controller.handleMockDelete)
}

func initAdminHostsController(r *gin.RouterGroup, controller *AdminHostsController) {
	r.GET("", controller.handleHostsConfigList)
	r.POST("", controller.handleHostConfigAddUpdate)

	r.GET("/:host", controller.handleHostConfigRetrieve)
	r.DELETE("/:host", controller.handleHostConfigDelete)

	r.POST("/:host/latencies", controller.handleLatencyAddUpdate)
	r.DELETE("/:host/latencies", controller.handleLatencyDelete)

	r.POST("/:host/errors", controller.handleErrorsAddUpdate)
	r.DELETE("/:host/errors/:error", controller.handleErrorDelete)

	r.POST("/:host/uris", controller.handleUrisAddUpdate)
}
