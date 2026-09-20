.PHONY: up down seed test test-integration coverage mocks lint

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

seed:
	docker exec -i agnos-db-1 psql \
		-U $$(grep ^DB_USER .env | cut -d= -f2) \
		-d $$(grep ^DB_NAME .env | cut -d= -f2) \
		< scripts/seed_patients.sql
