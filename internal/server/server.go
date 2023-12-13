package server

import (
	"fmt"

	"github.com/Caik/go-mock-server/internal/config"
	"github.com/Caik/go-mock-server/internal/server/controller"
	"github.com/Caik/go-mock-server/internal/server/middleware"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func Init() error {
	appConfig, err := config.GetAppConfig()

	if err != nil {
		return fmt.Errorf("error while getting app config: %v", err)
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Uuid)
	r.Use(middleware.Logger)

	controller.Init(r)

	log.Info(fmt.Sprintf("starting server on port %d", appConfig.ServerPort))

	if err := r.Run(fmt.Sprintf(":%d", appConfig.ServerPort)); err != nil {
		return err
	}

	return nil
}
