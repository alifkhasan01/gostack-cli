APP_NAME={{PROJECT_NAME}}
BIN_DIR=./bin

.PHONY: run build tidy test migrate-up migrate-down

run:
	go run ./cmd/api

build:
	go build -o $(BIN_DIR)/$(APP_NAME) ./cmd/api

tidy:
	go mod tidy

test:
	go test ./... -v

migrate-up:
	goose postgres "$(DB_DSN)" up -dir ./migrations

migrate-down:
	goose postgres "$(DB_DSN)" down -dir ./migrations
