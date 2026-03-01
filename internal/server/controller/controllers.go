package controller

import (
	"net/http"

	"github.com/Caik/go-mock-server/internal/rest"
	"github.com/gin-gonic/gin"
)

// InitMockRoutes initializes routes for the mock server
func InitMockRoutes(r *gin.Engine, mocksController *MocksController) {
	r.NoRoute(mocksController.handleMockRequest)
}

// InitAdminRoutes initializes routes for the admin server
func InitAdminRoutes(r *gin.Engine, adminMocksController *AdminMocksController, adminHostsController *AdminHostsController) {
	// Health check endpoint
	r.GET("/health", handleHealthCheck)

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		initAdminMocksController(v1.Group("/mocks"), adminMocksController)
		initAdminHostsController(v1.Group("/config/hosts"), adminHostsController)
	}
}

func handleHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, rest.Response{
		Status:  rest.Success,
		Message: "healthy",
	})
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
