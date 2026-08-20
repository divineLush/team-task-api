.PHONY: help up down restart logs build run test swagger migrate-up migrate-down clean

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

test: ## run all tests
	go test ./... -count=1

test-v: ## run all tests (verbose)
	go test ./... -v -count=1

test-race: ## run all tests with race detector
	go test ./... -race -count=1

test-cover: ## run tests with coverage report
	go test ./... -cover -count=1

test-cover-html: ## generate coverage report and open in browser
	go test ./... -coverprofile=coverage.out -count=1
	go tool cover -html=coverage.out
	rm coverage.out

test-pkg: ## run tests for a specific package (PKG=internal/service)
	go test ./$(PKG)/... -v -count=1

swagger: ## generate swagger docs
	swag init -g cmd/server/main.go

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
	rm -rf bin/ coverage.out
