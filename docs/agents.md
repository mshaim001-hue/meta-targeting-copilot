# Agents

Specialized services, not one giant prompt.

| Agent | Type | Output |
|---|---|---|
| Brief Analyzer | LLM | business model, goal, constraints, ≤5 questions |
| Audience Segmentation | LLM | 3–7 segments, no Meta API |
| Signal Generator | LLM | searchable phrases |
| Meta Resolver | deterministic | verified / suggested / unavailable / restricted |
| Geo Strategy | LLM + Meta geo search | radius, age, verified location keys |
| Retargeting Planner | deterministic from assets | custom audience tree |
| Offer Strategist | LLM | one offer per segment |
| Creative Strategist | LLM | one primary creative per segment |
| Strategy Critic | rules + LLM | overlapping segments, hallucinations, weak offers |

## What they must not do

- Invent Meta interest IDs
- Mark a signal as available because it “sounds like” a Meta interest
- Generate reach numbers
- Launch campaigns, ad sets, or creatives in Ads Manager (out of MVP)

## Prompts

Live in `backend/internal/ai/prompts.go`. Every call asks for JSON only. The engine unmarshals into Go structs and validates required fields before the next stage.
