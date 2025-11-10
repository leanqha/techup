.PHONY: build up down logs restart migrate psql

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

migrate:
	docker compose -f docker-compose.yml exec app goose -dir $${GOOSE_MIGRATION_DIR} $${GOOSE_DRIVER} "$${GOOSE_DBSTRING}"

psql:
	docker compose -f docker-compose.yml exec db psql -U $${DB_USER} -d $${DB_NAME}