package main

import (
	"github.com/Caik/go-mock-server/internal/ci"
	"github.com/Caik/go-mock-server/internal/config"
	"github.com/Caik/go-mock-server/internal/server"
	"github.com/Caik/go-mock-server/internal/server/controller"
	"github.com/Caik/go-mock-server/internal/service/admin"
	"github.com/Caik/go-mock-server/internal/service/cache"
	"github.com/Caik/go-mock-server/internal/service/content"
	"github.com/Caik/go-mock-server/internal/service/mock"
	"github.com/rs/zerolog/log"
	"go.uber.org/dig"
)

func main() {
	config.InitLogger()

	log.Info().
		Str("version", config.GetVersion()).
		Msg("starting mock server")

	if errs := setupCI(); len(errs) > 0 {
		log.Fatal().
			Msgf("error while setting up CI config: %v", errs)
	}

	// starting server
	if err := startServer(); err != nil {
		log.Fatal().
			Err(err).
			Stack().
			Msg("error while starting the server")
	}

	log.Info().
		Msg("shutting down the app")
}

func setupCI() []error {
	errs := make([]error, 0)

	// config
	if err := ci.Add(config.ParseAppArguments); err != nil {
		errs = append(errs, err)
	}

	if err := ci.Add(config.NewHostsConfig); err != nil {
		errs = append(errs, err)
	}

	if err := ci.Add(config.NewMocksDirectoryConfig); err != nil {
		errs = append(errs, err)
	}

	// server
	if err := ci.Add(server.NewServer); err != nil {
		errs = append(errs, err)
	}

	// controllers
	if err := ci.Add(controller.NewMocksController); err != nil {
		errs = append(errs, err)
	}

	if err := ci.Add(controller.NewAdminHostsController); err != nil {
		errs = append(errs, err)
	}

	if err := ci.Add(controller.NewAdminMocksController); err != nil {
		errs = append(errs, err)
	}

	// admin services
	if err := ci.Add(admin.NewHostsConfigAdminService); err != nil {
		errs = append(errs, err)
	}

	if err := ci.Add(admin.NewMockAdminService); err != nil {
		errs = append(errs, err)
	}

	// content services
	if err := ci.Add(content.NewFilesystemContentService, dig.As(new(content.ContentService))); err != nil {
		errs = append(errs, err)
	}

	// mock services
	if err := ci.Add(mock.NewMockServiceFactory); err != nil {
		errs = append(errs, err)
	}

	// cache service
	if err := ci.Add(cache.NewInMemoryCacheService, dig.As(new(cache.CacheService))); err != nil {
		errs = append(errs, err)
	}

	return errs
}

func startServer() error {
	return ci.Invoke(server.StartServer)
}
