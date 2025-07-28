# Usage:
# make run_docker                       # run docker environment

.PHONY: run_docker test

test:
	@echo ""
	@echo "########################################"
	@echo "##        Running all tests           ##"
	@echo "########################################"
	@echo ""
	@CGO_ENABLED=0 go test -timeout 30s ./internal/...

run_docker: ./docker-compose.yml
	@echo ""
	@echo "########################################"
	@echo "##     Running docker environment     ##"
	@echo "########################################"
	@echo ""
	@docker compose -f $< up --build --force-recreate