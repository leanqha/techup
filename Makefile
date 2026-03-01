.PHONY: build up down logs restart migrate migrate-up migrate-down migrate-status migrate-version psql test cover tidy

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
	docker compose -f docker-compose.yml exec app goose -dir $${GOOSE_MIGRATION_DIR} $${GOOSE_DRIVER} "$${GOOSE_DBSTRING}" up

migrate-down:
	docker compose -f docker-compose.yml exec app goose -dir $${GOOSE_MIGRATION_DIR} $${GOOSE_DRIVER} "$${GOOSE_DBSTRING}" down

migrate-status:
	docker compose -f docker-compose.yml exec app goose -dir $${GOOSE_MIGRATION_DIR} $${GOOSE_DRIVER} "$${GOOSE_DBSTRING}" status

migrate-version:
	docker compose -f docker-compose.yml exec app goose -dir $${GOOSE_MIGRATION_DIR} $${GOOSE_DRIVER} "$${GOOSE_DBSTRING}" version

psql:
	docker compose -f docker-compose.yml exec db psql -U $${DB_USER} -d $${DB_NAME}

test:
	go test ./... -p 1 -v

cover:
	go tool cover -html=coverage.out

tidy:
	go mod tidy