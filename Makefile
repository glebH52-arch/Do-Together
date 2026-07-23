include .env
export

service-run:
	@go run ./cmd/main.go 

migrate-up:
	@migrate -path ./migrations -database ${DATABASE_URL} up

migrate-down:
	@migrate -path ./migrations -database ${DATABASE_URL} down

