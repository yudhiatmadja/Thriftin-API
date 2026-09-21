.PHONY: run migrate-up migrate-down swag docker-up docker-down tidy
run:
	go run ./cmd/server
migrate-up:
	migrate -path migrations -database "\" up
migrate-down:
	migrate -path migrations -database "\" down 1
swag:
	swag init -g cmd/server/main.go -o docs
docker-up:
	docker compose up --build -d
docker-down:
	docker compose down -v
tidy:
	go mod tidy && go vet ./...
