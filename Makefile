# Usage:
# make                                  # build go source code, build go source code for windows
# make build                            # build go source code for current platform
# make test                             # run all unit tests						
# make build_mac_amd64                  # build go source code for mac AMD64
# make build_mac_arm64                  # build go source code for mac ARM64
# make build_linux                      # build go source code for linux
# make build_windows                    # build go source code for windows
# make build_and_push_docker_image      # build and push docker image
# make run_docker                       # run docker environment

all: test build_linux build_mac_amd64 build_mac_arm64 build_windows
.PHONY: all build test build_mac_amd64 build_mac_arm64 build_linux build_windows build_and_push_docker_image run_docker

build: ./cmd/mock-server/main.go
	@echo ""
	@echo "########################################"
	@echo "##       Building Mock Server         ##"
	@echo "########################################"
	@echo ""
	@CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static" -s -w' -o dist/mock-server $<

test:
	@echo ""
	@echo "########################################"
	@echo "##        Running all tests           ##"
	@echo "########################################"
	@echo ""
	@CGO_ENABLED=0 go test -timeout 30s ./internal/...

build_mac_amd64: ./cmd/mock-server/main.go
	@echo ""
	@echo "########################################"
	@echo "## Building Mock Server for Mac AMD64 ##"
	@echo "########################################"
	@echo ""
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -ldflags '-extldflags "-static" -s -w' -o dist/mock-server_mac_amd64 $<

build_mac_arm64: ./cmd/mock-server/main.go
	@echo ""
	@echo "########################################"
	@echo "## Building Mock Server for Mac ARM64 ##"
	@echo "########################################"
	@echo ""
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -a -ldflags '-extldflags "-static" -s -w' -o dist/mock-server_mac_arm64 $<

build_linux: ./cmd/mock-server/main.go
	@echo ""
	@echo "########################################"
	@echo "##  Building Mock Server for Linux    ##"
	@echo "########################################"
	@echo ""
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-extldflags "-static" -s -w' -o dist/mock-server_linux $<

build_windows: ./cmd/mock-server/main.go
	@echo ""
	@echo "########################################"
	@echo "##  Building Mock Server for Windows  ##"
	@echo "########################################"
	@echo ""
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -a -ldflags '-extldflags "-static" -s -w' -o dist/mock-server.exe $<

build_and_push_docker_image: ./Dockerfile
	@echo ""
	@echo "########################################"
	@echo "##       Building docker image        ##"
	@echo "########################################"
	@echo ""
	@docker buildx build -f $< -t caik/go-mock-server:latest --push --platform=linux/amd64,linux/arm64,linux/arm/v7 .

run_docker: ./docker-compose.yml
	@echo ""
	@echo "########################################"
	@echo "##     Running docker environment     ##"
	@echo "########################################"
	@echo ""
	@docker compose -f $< up --build --force-recreate