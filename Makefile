.PHONY: swagger migrate setup

swagger:
	swag init -g cmd/main.go -d .

migrate:
	go run cmd/migrate/main.go

setup: swagger migrate
