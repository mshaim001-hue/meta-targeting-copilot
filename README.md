# AI Targeting Copilot for Meta Ads

Turns a natural-language audience description into a **grounded** Meta Ads targeting strategy.

LLM does the marketing thinking. Meta API decides what is actually targetable.

## Stack

- Frontend: React + TypeScript + Vite + Tailwind + shadcn-style UI
- Backend: Go, chi, JWT
- DB: PostgreSQL, pgx (`sqlc` query files included)
- Jobs: table queue in Postgres
- LLM: `LLMProvider` (OpenAI-compatible, mock fallback)
- Meta: `internal/meta` Graph adapter (mock catalog fallback)

## Quick start

```bash
cp .env.example .env
docker compose up db -d

# API
cd backend && go run ./cmd/api

# UI (other terminal)
cd frontend && npm install && npm run dev
```

Open [http://localhost:5173](http://localhost:5173). Register, create a project, paste an audience prompt, click **Build Strategy**.

Without `LLM_API_KEY` / `META_ACCESS_TOKEN` the app still runs on mocks so you can walk the full flow. Connect real keys to verify against Graph API v26.

## Pipeline CLI (milestone 1)

```bash
cd backend
go run ./cmd/pipeline \
  --product "Fitness club" \
  --city Astana \
  --country Kazakhstan \
  --audience "People who already exercise regularly."
```

## Tests

```bash
cd backend && go test ./...
```

## Out of MVP

No campaign creation, no ad set push, no billing, no image/video generation, no Ads Library scraping.

## Docs

- [Architecture](docs/architecture.md)
- [API](docs/api.md)
- [Agents](docs/agents.md)
