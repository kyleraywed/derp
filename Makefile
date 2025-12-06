default:
	@echo "Use: make run OR make build"

run:
	@go mod tidy
	@go run cmd/main.go

build:
	@go mod tidy
	@go build -o bin/dee cmd/main.go