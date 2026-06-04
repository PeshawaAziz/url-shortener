# Makefile

.PHONY: run build migrate-up migrate-down docker-up docker-down test lint

# ─── Variables ───────────────────────────────────────────────
APP_NAME        = url-shortener
CMD_PATH        = ./cmd/api
MIGRATE_CMD     = ./cmd/migrate
MIGRATIONS_PATH = ./migrations
DB_URL          = postgresql://postgres:postgres@localhost:5432/urlshortener?sslmode=disable

# ─── Development ─────────────────────────────────────────────
run:
	go run $(CMD_PATH)/main.go

build:
	go build -o bin/$(APP_NAME) $(CMD_PATH)/main.go

# ─── Database ────────────────────────────────────────────────
migrate-up:
	go run $(MIGRATE_CMD)/main.go up

migrate-down:
	go run $(MIGRATE_CMD)/main.go down

migrate-create:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $$name

# ─── Docker ──────────────────────────────────────────────────
docker-up:
	docker compose -f docker/docker-compose.yml up -d

docker-down:
	docker compose -f docker/docker-compose.yml down

docker-build:
	docker compose -f docker/docker-compose.yml up -d --build

# ─── Quality ─────────────────────────────────────────────────
test:
	go test ./... -v -race -count=1

test-cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out

lint:
	golangci-lint run ./...

# ─── Helpers ─────────────────────────────────────────────────
tidy:
	go mod tidy

.DEFAULT_GOAL := run