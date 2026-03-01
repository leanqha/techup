.PHONY: build up down logs restart migrate migrate-up migrate-down migrate-status migrate-version psql test cover tidy

ENV_FILE ?= ./.env

build:
	docker compose -f docker-compose.yml build

up:
	docker compose -f docker-compose.yml up -d
	docker compose -f docker-compose.yml logs -f

down:
	docker compose -f docker-compose.yml down -v

logs:
	docker compose -f docker-compose.yml logs -f app

restart:
	docker compose -f docker-compose.yml down -v
	docker compose -f docker-compose.yml build
	docker compose -f docker-compose.yml up -d
	docker compose -f docker-compose.yml logs -f

migrate: migrate-up

migrate-up:
	sh -c 'export $$(grep -E "^[A-Za-z_][A-Za-z0-9_]*=" $(ENV_FILE) | sed -E "s/[[:space:]]+#.*$$//" | xargs); GOOSE_DRIVER=$${GOOSE_DRIVER:-postgres} GOOSE_DBSTRING="user=$$DB_USER password=$$DB_PASSWORD host=$$DB_HOST port=$$DB_PORT dbname=$$DB_NAME sslmode=$$DB_SSLMODE" goose -env $(ENV_FILE) -dir "$${GOOSE_MIGRATION_DIR:-./migrations}" up'

migrate-down:
	sh -c 'export $$(grep -E "^[A-Za-z_][A-Za-z0-9_]*=" $(ENV_FILE) | sed -E "s/[[:space:]]+#.*$$//" | xargs); GOOSE_DRIVER=$${GOOSE_DRIVER:-postgres} GOOSE_DBSTRING="user=$$DB_USER password=$$DB_PASSWORD host=$$DB_HOST port=$$DB_PORT dbname=$$DB_NAME sslmode=$$DB_SSLMODE" goose -env $(ENV_FILE) -dir "$${GOOSE_MIGRATION_DIR:-./migrations}" down'

migrate-status:
	sh -c 'export $$(grep -E "^[A-Za-z_][A-Za-z0-9_]*=" $(ENV_FILE) | sed -E "s/[[:space:]]+#.*$$//" | xargs); GOOSE_DRIVER=$${GOOSE_DRIVER:-postgres} GOOSE_DBSTRING="user=$$DB_USER password=$$DB_PASSWORD host=$$DB_HOST port=$$DB_PORT dbname=$$DB_NAME sslmode=$$DB_SSLMODE" goose -env $(ENV_FILE) -dir "$${GOOSE_MIGRATION_DIR:-./migrations}" status'

migrate-version:
	sh -c 'export $$(grep -E "^[A-Za-z_][A-Za-z0-9_]*=" $(ENV_FILE) | sed -E "s/[[:space:]]+#.*$$//" | xargs); GOOSE_DRIVER=$${GOOSE_DRIVER:-postgres} GOOSE_DBSTRING="user=$$DB_USER password=$$DB_PASSWORD host=$$DB_HOST port=$$DB_PORT dbname=$$DB_NAME sslmode=$$DB_SSLMODE" goose -env $(ENV_FILE) -dir "$${GOOSE_MIGRATION_DIR:-./migrations}" version'

psql:
	docker compose -f docker-compose.yml exec db psql -U $${DB_USER} -d $${DB_NAME}

test:
	go test ./... -p 1 -v

cover:
	go tool cover -html=coverage.out

tidy:
	go mod tidy