.PHONY: run ingest tidy lint test migrate-up migrate-down

run:
	go run ./cmd/api

ingest:
	go run ./cmd/ingest

migrate-up:
	migrate -path migrations -database $$DATABASE_URL up

migrate-down:
	migrate -path migrations -database $$DATABASE_URL down

tidy:
	go mod tidy

lint:
	golangci-lint run

test:
	go test ./...
