STACK ?= grafana-lgtm
COMPOSE := docker compose -f stacks/$(STACK)/docker-compose.yml -p binoc

.PHONY: up down logs build

up: ## Start the stack (default: grafana-lgtm)
	$(COMPOSE) up --build -d

down: ## Stop the stack
	$(COMPOSE) down

logs: ## Tail logs from all services
	$(COMPOSE) logs -f

build: ## Build the echo service image
	$(COMPOSE) build echo
