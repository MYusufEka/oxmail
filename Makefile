.PHONY: up down build test test-go test-web test-e2e lint logs dev seed reset clean help env prod prod-down prod-test

# Default target
.DEFAULT_GOAL := help

# =============================================================================
# Oxmail Makefile
# =============================================================================

## up: Start all services in production mode
up:
	docker compose up -d

## down: Stop all services
down:
	docker compose down

## build: Build all images
build:
	docker compose build

## test: Run all tests (Go + Vitest + Playwright)
test: test-go test-web test-e2e

## test-go: Run Go tests
test-go:
	cd cmd/oxmail-api && go test ./...
	cd cmd/oxmail && go test ./...
	go test ./internal/...

## test-web: Run Vitest (frontend unit tests)
test-web:
	cd web && npm test

## test-e2e: Run E2E integration tests
test-e2e:
	./scripts/test-e2e.sh

## lint: Run all linters
lint:
	golangci-lint run ./...
	cd web && npm run lint

## logs: Tail all container logs
logs:
	docker compose logs -f

## env: Create .env from .env.example (if not exists)
env:
	@test -f .env || (cp .env.example .env && echo "Created .env from .env.example")
	@test -f .env && echo ".env already exists" || true

## dev: Start full stack in dev mode
dev: env
	docker compose -f docker-compose.yml -f docker-compose.dev.yml up -d --build

## seed: Seed test data (domain + users + emails)
seed:
	./scripts/seed.sh

## reset: Wipe and re-seed
reset:
	docker compose down -v
	docker compose -f docker-compose.yml -f docker-compose.dev.yml up -d --build
	./scripts/seed.sh

## prod: Start production stack (Traefik + TLS + production ports)
prod: env
	docker compose --profile prod -f docker-compose.yml -f docker-compose.prod.yml up -d --build

## prod-down: Stop production stack
prod-down:
	docker compose --profile prod -f docker-compose.yml -f docker-compose.prod.yml down

## prod-test: Run production integration tests
prod-test:
	./tests/e2e/production_test.sh

## clean: Remove all containers, volumes, and build artifacts
clean:
	docker compose down -v --rmi local
	rm -rf tmp/

## help: Show this help message
help:
	@echo "Oxmail - Dockerized Mail Server"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@grep -E '^## [a-zA-Z_-]+:' $(MAKEFILE_LIST) | awk -F ':' '{printf "  \033[36m%-15s\033[0m %s\n", $$2, $$3}' | sed 's/^ *//'
