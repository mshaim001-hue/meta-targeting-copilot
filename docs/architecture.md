# Architecture

The TZ recommended Next.js + Supabase. This repo follows the chosen split instead:

- **Frontend:** React + TypeScript + Vite + Tailwind + shadcn-style UI
- **Backend:** Go, chi, PostgreSQL, pgx
- **Auth:** JWT issued by the Go API
- **LLM:** `internal/ai.Provider` (OpenAI-compatible or mock)
- **Meta:** `internal/meta.Provider` (Graph API or mock catalog)
- **Jobs:** PostgreSQL queue with `FOR UPDATE SKIP LOCKED`

## Data flow

```text
React
  ↓  POST /api/projects/:id/analyze
Go API  →  jobs table
  ↓
Brief Analyzer          (LLM)
  ↓
Audience Segmenter      (LLM)
  ↓
Signal Generator        (LLM)
  ↓
Meta Resolver           (Meta API, never LLM)
  ↓
Strategy Engine         (deterministic)
  ↓
Offer / Creative Engine (LLM)
  ↓
React
```

## Grounding rule

LLM output is marketing hypothesis. It cannot become a `TargetingSpec`.

```text
LLM suggestion → validator → Meta resolver → verified targeting → TargetingSpec
```

A UI chip is `Meta verified` only when **both** are true:

1. `source = meta_verified`
2. `metaId` came from the Meta provider

## Packages

| Package | Responsibility |
|---|---|
| `internal/ai` | Structured JSON from an LLM. No Meta IDs. |
| `internal/meta` | Interests, geo, suggestions, reach estimate. Version via `META_API_VERSION`. |
| `internal/targeting` | Resolve signals, strip hallucinations, score. |
| `internal/strategy` | Broad / Interest / Lookalike / Retargeting. Export JSON+Markdown. |
| `internal/pipeline` | Orchestrates the stages from the TZ. |
| `internal/jobs` | DB-backed worker loop inside the API process. |
| `internal/db` | Migrations + pgx store. `sqlc` queries live in `internal/db/queries` for the next generation pass. |

## Providers

If `LLM_API_KEY` is empty, the mock LLM runs so the pipeline is demoable.

If `META_ACCESS_TOKEN` is empty, a static interest/geo catalog is used. The UI labels this as mock. Mock estimates are **never** presented as Graph API numbers in copy — the strategy JSON still flags `metaMode`.

## First milestone

Prove the engine before extra UI:

```bash
cd backend
go run ./cmd/pipeline --product "Fitness club" --city Astana --country Kazakhstan \
  --audience "People who already exercise regularly."
```
