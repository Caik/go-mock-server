package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/Caik/go-mock-server/internal/server/controller"
	"go.uber.org/dig"

	"github.com/Caik/go-mock-server/internal/config"
	"github.com/Caik/go-mock-server/internal/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type Servers struct {
	MockEngine  *gin.Engine
	AdminEngine *gin.Engine
}

type StartServerParams struct {
	dig.In

	Servers              *Servers
	AppArguments         *config.AppArguments
	AdminMocksController *controller.AdminMocksController
	AdminHostsController *controller.AdminHostsController
	TrafficController    *controller.TrafficController
	MocksController      *controller.MocksController
}

var once sync.Once

func NewServers() *Servers {
	once.Do(func() {
		gin.SetMode(gin.ReleaseMode)
	})

	return &Servers{
		MockEngine:  newGinEngine(),
		AdminEngine: newAdminGinEngine(),
	}
}

func newGinEngine() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Uuid)
	r.Use(middleware.Logger)

	return r
}

func newAdminGinEngine() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	// Enable CORS only in dev mode
	if config.GetVersion() == "dev" {
		r.Use(middleware.Cors)
	}

	r.Use(middleware.Uuid)
	r.Use(middleware.Logger)

	return r
}

func StartServers(params StartServerParams) error {
	// Initialize mock routes on mock engine
	controller.InitMockRoutes(params.Servers.MockEngine, params.MocksController)

	// Initialize admin routes on admin engine
	controller.InitAdminRoutes(params.Servers.AdminEngine, params.AdminMocksController, params.AdminHostsController, params.TrafficController)

	// Initialize UI routes if a UI directory is configured
	if params.AppArguments.UIDirectory != "" {
		controller.InitUIRoutes(params.Servers.AdminEngine, params.AppArguments.UIDirectory)
	}

	// Channel to capture errors from goroutines
	errChan := make(chan error, 2)

	// Context for coordinating shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start admin server if enabled
	if params.AppArguments.AdminPort > 0 {
		go func() {
			log.Info().
				Msgf("starting admin server on port %d", params.AppArguments.AdminPort)

			server := &http.Server{
				Addr:    fmt.Sprintf(":%d", params.AppArguments.AdminPort),
				Handler: params.Servers.AdminEngine,
			}

			go func() {
				<-ctx.Done()
				err := server.Shutdown(context.Background())

				if err != nil {
					log.Err(err).
						Stack().
						Msg("error while shutting down admin server")
				}
			}()

			if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				errChan <- fmt.Errorf("admin server error: %w", err)
				cancel()
			}
		}()
	}

	// Start mock server
	log.Info().
		Msgf("starting mock server on port %d", params.AppArguments.ServerPort)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", params.AppArguments.ServerPort),
		Handler: params.Servers.MockEngine,
	}

	go func() {
		<-ctx.Done()
		err := server.Shutdown(context.Background())

		if err != nil {
			log.Err(err).
				Stack().
				Msg("error while shutting down mock server")
		}
	}()

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("mock server error: %w", err)
	}

	// Check if admin server had an error
	select {
	case err := <-errChan:
		return err
	default:
		return nil
	}
}
