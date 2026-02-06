.PHONY: dev build test docker-up docker-down logs

dev:
	cd server && go run ./cmd/server

build:
	cd server && go build -o bin/server ./cmd/server

test:
	cd server && go test -v ./...

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

logs:
	docker-compose logs -f server
