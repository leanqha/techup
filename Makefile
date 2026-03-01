.PHONY: build up down logs restart migrate migrate-up migrate-down migrate-status migrate-version psql test cover tidy

MIGRATIONS_DIR ?= /app/migrations

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
	sh -c 'set -a; . ./.env; set +a; goose -dir "${GOOSE_MIGRATION_DIR:-./migrations}" "$$GOOSE_DRIVER" "user=$$DB_USER password=$$DB_PASSWORD host=$$DB_HOST port=$$DB_PORT dbname=$$DB_NAME sslmode=$$DB_SSLMODE" up'

migrate-down:
	sh -c 'set -a; . ./.env; set +a; goose -dir "${GOOSE_MIGRATION_DIR:-./migrations}" "$$GOOSE_DRIVER" "user=$$DB_USER password=$$DB_PASSWORD host=$$DB_HOST port=$$DB_PORT dbname=$$DB_NAME sslmode=$$DB_SSLMODE" down'

migrate-status:
	sh -c 'set -a; . ./.env; set +a; goose -dir "${GOOSE_MIGRATION_DIR:-./migrations}" "$$GOOSE_DRIVER" "user=$$DB_USER password=$$DB_PASSWORD host=$$DB_HOST port=$$DB_PORT dbname=$$DB_NAME sslmode=$$DB_SSLMODE" status'

migrate-version:
	sh -c 'set -a; . ./.env; set +a; goose -dir "${GOOSE_MIGRATION_DIR:-./migrations}" "$$GOOSE_DRIVER" "user=$$DB_USER password=$$DB_PASSWORD host=$$DB_HOST port=$$DB_PORT dbname=$$DB_NAME sslmode=$$DB_SSLMODE" version'

psql:
	docker compose -f docker-compose.yml exec db psql -U $${DB_USER} -d $${DB_NAME}

test:
	go test ./... -p 1 -v

cover:
	go tool cover -html=coverage.out

tidy:
	go mod tidy