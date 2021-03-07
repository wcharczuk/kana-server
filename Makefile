run:
	@go run main.go

run-reload:
	@reflex -s -- go run main.go

db:
	@go run main.go --create-database=true --migrations=true --server=false

migrate:
	@go run main.go --create-database=false --migrations=true --server=false
