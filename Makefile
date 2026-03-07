# Usage:
# make test          # run Go tests
# make build-ui      # build the React frontend
# make run_docker    # run docker environment

.PHONY: run_docker test build-ui

test:
	@echo ""
	@echo "########################################"
	@echo "##        Running all tests           ##"
	@echo "########################################"
	@echo ""
	@CGO_ENABLED=0 go test -timeout 30s ./internal/...

build-ui:
	@echo ""
	@echo "########################################"
	@echo "##       Building frontend UI         ##"
	@echo "########################################"
	@echo ""
	@cd web && npm ci && npm run build

run_docker: ./docker-compose.yml
	@echo ""
	@echo "########################################"
	@echo "##     Running docker environment     ##"
	@echo "########################################"
	@echo ""
	@docker compose -f $< up --build --force-recreate