.PHONY: up down test test-integration coverage mocks lint

up:
	docker compose up --build -d

down:
	docker compose down

test:
	go test ./... -cover

test-integration:
	go test -tags=integration ./... -cover

coverage:
	go test -tags=integration ./... -coverprofile=coverage.out && go tool cover -html=coverage.out

mocks:
	mockery

lint:
	go vet ./...
