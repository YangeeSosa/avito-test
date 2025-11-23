.PHONY: run build test lint docker

run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

test:
	go test ./...

docker:
	docker-compose up --build
