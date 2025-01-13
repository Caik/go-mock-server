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
	"go.uber.org/dig"

	log "github.com/sirupsen/logrus"
)

func main() {
	config.InitLogger()

	log.WithField("version", config.GetVersion()).
		Info("starting mock server")

	if errs := setupCI(); len(errs) > 0 {
		log.Fatalf("error while setting up CI config: %v", errs)
	}

	// start server
	if err := startServer(); err != nil {
		log.Fatalf("error while starting the server: %v", err)
	}

	log.Info("shutting down the app")
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
