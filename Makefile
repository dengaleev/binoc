STACK ?= loki-tempo-prometheus
COMPOSE := docker compose -f stacks/$(STACK)/docker-compose.yml -p binoc-$(STACK)

.PHONY: up down logs build list

up: ## Start the stack (default: loki-tempo-prometheus)
	$(COMPOSE) up --build -d

down: ## Stop the stack and remove volumes
	$(COMPOSE) down -v

logs: ## Tail logs from all services
	$(COMPOSE) logs -f

build: ## Build the service image
	$(COMPOSE) build app

list: ## List available stacks
	@ls -1 stacks/
