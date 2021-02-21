run:
	@go run main.go

run-reload:
	@reflex -s -- go run main.go

db:
	@go run migrations/main.go
