.PHONY: dev build test web-dev web-build openclaw-build openclaw-agent \
	docker-up docker-up-build docker-down logs logs-openclaw logs-all

COMPOSE := $(shell if command -v docker-compose >/dev/null 2>&1; then echo docker-compose; else echo "docker compose"; fi)

dev:
	cd server && go run ./cmd/server

build:
	cd server && go build -o bin/server ./cmd/server

test:
	cd server && go test -v ./...

web-dev:
	cd web && npm run dev

web-build:
	cd web && npm run build

openclaw-build:
	cd openclaw-skill && npm run build

openclaw-agent:
	cd openclaw-skill && openclaw agent --local --session-id easymeme --message "获取待分析代币 -> AI 分析 -> 回写结果"

docker-up:
	$(COMPOSE) up -d

docker-up-build:
	$(COMPOSE) up --build

docker-down:
	$(COMPOSE) down

logs:
	$(COMPOSE) logs -f server

logs-openclaw:
	$(COMPOSE) logs -f openclaw

logs-all:
	$(COMPOSE) logs -f
