include .env
export

service-run:
	@go run ./cmd/main.go 

migrate-up:
	@migrate -path ./migrations -database ${connectionString} up

migrate-down:
	@migrate -path ./migrations -database ${connectionString} down