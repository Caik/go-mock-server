# Usage:
# make                                  # build go source code, build go source code for windows
# make build_mac                        # build go source code for mac
# make build_linux                      # build go source code for linux
# make build_windows                    # build go source code for windows
# make push_docker_image                # push docker image to DockerHub
# make build_docker_image               # build docker image binary
# make run_docker                       # run docker environment

all: build_linux build_mac build_windows
.PHONY: all build_mac build_linux build_windows build_docker_image push_docker_image run_docker

build_mac: ./cmd/mock-server/main.go
	@echo ""
	@echo "######################################"
	@echo "##   Building Mock Server for Mac   ##"
	@echo "######################################"
	@echo ""
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -ldflags '-extldflags "-static" -s -w' -o dist/mock-server_mac $<

build_linux: ./cmd/mock-server/main.go
	@echo ""
	@echo "######################################"
	@echo "##  Building Mock Server for Linux  ##"
	@echo "######################################"
	@echo ""
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-extldflags "-static" -s -w' -o dist/mock-server_linux $<

build_windows: ./cmd/mock-server/main.go
	@echo ""
	@echo "######################################"
	@echo "## Building Mock Server for Windows ##"
	@echo "######################################"
	@echo ""
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -a -ldflags '-extldflags "-static" -s -w' -o dist/mock-server.exe $<

build_docker_image: ./build/docker/Dockerfile
	@echo ""
	@echo "######################################"
	@echo "##      Building docker image       ##"
	@echo "######################################"
	@echo ""
	@docker build -f $< -t caik/go-mock-server:latest .

push_docker_image: ./build/docker/Dockerfile
	@echo ""
	@echo "######################################"
	@echo "##      Pushing docker image        ##"
	@echo "######################################"
	@echo ""
	@docker push caik/go-mock-server:latest

run_docker: ./build/docker/docker-compose.yml
	@echo ""
	@echo "######################################"
	@echo "##    Running docker environment    ##"
	@echo "######################################"
	@echo ""
	@docker-compose -f $< up --build --force-recreate