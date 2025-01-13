package server

import (
	"fmt"
	"github.com/Caik/go-mock-server/internal/server/controller"
	"go.uber.org/dig"

	"github.com/Caik/go-mock-server/internal/config"
	"github.com/Caik/go-mock-server/internal/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type StartServerParams struct {
	dig.In

	Engine               *gin.Engine
	AppArguments         *config.AppArguments
	AdminMocksController *controller.AdminMocksController
	AdminHostsController *controller.AdminHostsController
	MocksController      *controller.MocksController
}

func NewServer() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Uuid)
	r.Use(middleware.Logger)

	return r
}

func StartServer(params StartServerParams) error {
	controller.InitRoutes(params.Engine, params.AdminMocksController, params.AdminHostsController, params.MocksController)

	log.Info().
		Msgf("starting server on port %d", params.AppArguments.ServerPort)

	if err := params.Engine.Run(fmt.Sprintf(":%d", params.AppArguments.ServerPort)); err != nil {
		return err
	}

	return nil
}
