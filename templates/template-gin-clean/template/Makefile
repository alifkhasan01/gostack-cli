APP_NAME={{PROJECT_NAME}}
BIN_DIR=./bin

.PHONY: run build tidy lint test clean migrate migrate-up migrate-down

run:
	go run ./cmd/api

build:
	go build -o $(BIN_DIR)/$(APP_NAME) ./cmd/api

tidy:
	go mod tidy

lint:
	golangci-lint run ./...

test:
	go test ./... -v

clean:
	rm -rf $(BIN_DIR)

migrate:
	goose postgres "$(DB_DSN)" up -dir ./migrations

migrate-up:
	goose postgres "$(DB_DSN)" up -dir ./migrations

migrate-down:
	goose postgres "$(DB_DSN)" down -dir ./migrations
