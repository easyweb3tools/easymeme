# EasyWeb3 Monorepo

EasyWeb3 is a multi-application Web3 platform monorepo.

This repository now separates shared infrastructure/services from application code:

- `infra/`: shared runtime infrastructure (PostgreSQL, Redis, Nginx, base-service)
- `services/base/`: reusable base-service (third-party integrations, cache, wallet/notification APIs)
- `packages/go-sdk/`: Go SDK for app-to-base-service calls
- `apps/easymeme/`: EasyMeme application (server/web/openclaw-skill)

## Repository Layout

```text
easyweb3/
├── infra/
├── services/
│   └── base/
├── packages/
│   └── go-sdk/
└── apps/
    └── easymeme/
        ├── server/
        ├── web/
        ├── openclaw-skill/
        ├── README.md
        └── README_CN.md
```

## Quick Start

1. Prepare env file:

```bash
cp infra/.env.example .env
```

2. Start infra + EasyMeme:

```bash
make dev
```

3. Start infra only:

```bash
make infra
```

4. Stop all:

```bash
make stop
```

## Build and Test

```bash
make build-all
make test-all
```

## Application Documentation

EasyMeme docs were moved to:

- English: `apps/easymeme/README.md`
- 中文: `apps/easymeme/README_CN.md`

## CI Workflow Note

GitHub workflow paths were updated to the monorepo layout:

- server image build context: `apps/easymeme/server`
- web image build context: `apps/easymeme/web`
- openclaw image build context: `apps/easymeme/openclaw-skill`
