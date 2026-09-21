# Thriftin API
Go 1.24 + Gin + pgx + Redis + JWT
## Run
cp .env.example .env
docker compose up --build -d
go run ./cmd/server
## Swagger
/sw/swagger/index.html
## Make
make swag && make tidy
