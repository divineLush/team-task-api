.PHONY: help up down restart logs build run test migrate-up migrate-down clean

DB_CONTAINER := team-task-mysql
DB_USER := teamtask
DB_PASS := teamtask
DB_NAME := teamtask

help: ## show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

up: ## start docker containers
	docker compose up -d

down: ## stop docker containers
	docker compose down

restart: ## restart docker containers
	docker compose restart

logs: ## tail docker logs
	docker compose logs -f

build: ## build the server binary
	go build -o bin/server ./cmd/server

run: ## run the server
	go run ./cmd/server

test: ## run tests
	go test ./... -v

migrate-up: ## run all migrations
	@for f in migrations/*.up.sql; do \
		echo "applying $$f"; \
		docker exec -i $(DB_CONTAINER) mysql -u$(DB_USER) -p$(DB_PASS) $(DB_NAME) < $$f; \
	done

migrate-down: ## rollback all migrations (reverse order)
	@for f in $$(ls -r migrations/*.down.sql); do \
		echo "rolling back $$f"; \
		docker exec -i $(DB_CONTAINER) mysql -u$(DB_USER) -p$(DB_PASS) $(DB_NAME) < $$f; \
	done

clean: ## remove build artifacts
	rm -rf bin/
