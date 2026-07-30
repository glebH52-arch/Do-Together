include .env
export

service-run:
	@go run ./cmd/main.go

migrate-up:
	@migrate -path ./migrations -database ${DATABASE_URL} up

migrate-down:
	@migrate -path ./migrations -database ${DATABASE_URL} down

migrate-force:
	@migrate -path ./migrations -database ${DATABASE_URL} force ${VERSION}

migrate-down-one:
	@migrate -path ./migrations -database ${DATABASE_URL} down 1

migrate-status:
	@migrate -path ./migrations -database ${DATABASE_URL} version
