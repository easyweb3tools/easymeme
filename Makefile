.PHONY: dev infra stop clean build-base build-easymeme build-all test-base test-easymeme test-all

dev:
	docker compose -f infra/docker-compose.yml -f apps/easymeme/docker-compose.yml up --build

infra:
	docker compose -f infra/docker-compose.yml up --build

stop:
	docker compose -f infra/docker-compose.yml -f apps/easymeme/docker-compose.yml down

clean:
	docker compose -f infra/docker-compose.yml -f apps/easymeme/docker-compose.yml down -v

build-base:
	cd services/base && go build ./cmd/server/

build-easymeme:
	cd apps/easymeme/server && go build ./cmd/server/

build-all: build-base build-easymeme

test-base:
	cd services/base && go test ./...

test-easymeme:
	cd apps/easymeme/server && go test ./...

test-all: test-base test-easymeme
