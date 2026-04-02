.PHONY: build run test test-cover lint docker-up docker-down smoke

build:
	go build -o gateway ./cmd/gateway

run:
	go run ./cmd/gateway

test:
	go test ./...

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

lint:
	golangci-lint run

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-up-db:
	docker compose up -d postgres

smoke:
	./scripts/smoke_chat.sh
