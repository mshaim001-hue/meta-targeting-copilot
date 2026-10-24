# HTTP API

Base URL: `http://localhost:8080`

Auth: `Authorization: Bearer <jwt>` on every `/api/*` route except register/login.

## Auth

```
POST /api/auth/register   { email, password }
POST /api/auth/login      { email, password }
GET  /api/auth/me
```

## Projects

```
POST /api/projects
GET  /api/projects
GET  /api/projects/:id
PUT  /api/projects/:id
POST /api/projects/:id/brief          { prompt }
POST /api/projects/:id/analyze        { prompt? }   → 202 job
POST /api/projects/:id/strategy/regenerate
GET  /api/projects/:id/jobs/:jobId
GET  /api/projects/:id/job
GET  /api/projects/:id/segments
GET  /api/projects/:id/strategy
GET  /api/projects/:id/export?format=json|md
```

Job stages:

```
QUEUED → ANALYZING_BRIEF → CREATING_SEGMENTS → GENERATING_SIGNALS
 → VERIFYING_META_TARGETING → BUILDING_STRATEGIES → BUILDING_CREATIVES → COMPLETE
```

## Meta utilities

```
POST /api/meta/interests/search   { query }
POST /api/meta/geo/search         { query }
POST /api/meta/audience/estimate  TargetingSpec
```

If Graph API cannot estimate, the body is `{ "available": false, "message": "Estimate unavailable" }`. The LLM is not asked for a number.

## Project payload

```json
{
  "businessName": "Apex Gym",
  "website": "https://example.com",
  "product": "Fitness club",
  "country": "Kazakhstan",
  "city": "Astana",
  "price": "45000 KZT",
  "objective": "Membership",
  "budget": "3000 USD",
  "assets": {
    "customerList": false,
    "pixel": true,
    "instagram": true
  }
}
```
