package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/Caik/go-mock-server/internal/config"
	"github.com/Caik/go-mock-server/internal/server"
)

func main() {
	if _, err := config.Init(); err != nil {
		log.Fatalf("error while initializing config: %v", err)
	}

	log.Infof("starting Go Mock Server, version: %s", config.GetVersion())

	if err := server.Init(); err != nil {
		log.Fatalf("error while initializing server: %v", err)
	}
}
