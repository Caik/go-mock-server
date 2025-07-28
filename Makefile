# Usage:
# make run_docker                       # run docker environment

.PHONY: run_docker

run_docker: ./docker-compose.yml
	@echo ""
	@echo "########################################"
	@echo "##     Running docker environment     ##"
	@echo "########################################"
	@echo ""
	@docker compose -f $< up --build --force-recreate