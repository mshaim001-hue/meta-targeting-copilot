CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    business_name TEXT NOT NULL,
    website TEXT NOT NULL DEFAULT '',
    product TEXT NOT NULL,
    country TEXT NOT NULL,
    city TEXT NOT NULL,
    price TEXT NOT NULL DEFAULT '',
    objective TEXT NOT NULL DEFAULT '',
    budget TEXT NOT NULL DEFAULT '',
    has_customer_list BOOLEAN NOT NULL DEFAULT false,
    has_pixel BOOLEAN NOT NULL DEFAULT false,
    has_capi BOOLEAN NOT NULL DEFAULT false,
    has_website_visitors BOOLEAN NOT NULL DEFAULT false,
    has_instagram BOOLEAN NOT NULL DEFAULT false,
    has_facebook_page BOOLEAN NOT NULL DEFAULT false,
    has_video_views BOOLEAN NOT NULL DEFAULT false,
    has_lead_forms BOOLEAN NOT NULL DEFAULT false,
    has_purchasers BOOLEAN NOT NULL DEFAULT false,
    has_crm BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE audience_briefs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    prompt TEXT NOT NULL,
    analyzed_json JSONB,
    missing_info JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE audience_segments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    brief_id UUID NOT NULL REFERENCES audience_briefs(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    intent_level TEXT NOT NULL,
    pains JSONB NOT NULL DEFAULT '[]'::jsonb,
    motivations JSONB NOT NULL DEFAULT '[]'::jsonb,
    semantic_signals JSONB NOT NULL DEFAULT '[]'::jsonb,
    offer_angles JSONB NOT NULL DEFAULT '[]'::jsonb,
    creative_angles JSONB NOT NULL DEFAULT '[]'::jsonb,
    position INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE targeting_entities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    segment_id UUID REFERENCES audience_segments(id) ON DELETE CASCADE,
    brief_id UUID NOT NULL REFERENCES audience_briefs(id) ON DELETE CASCADE,
    query TEXT NOT NULL,
    meta_id TEXT,
    meta_name TEXT,
    meta_type TEXT,
    status TEXT NOT NULL,
    source TEXT NOT NULL,
    generated_from TEXT NOT NULL DEFAULT '',
    explanation TEXT NOT NULL DEFAULT '',
    audience_size BIGINT,
    raw JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE audience_strategies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    brief_id UUID NOT NULL REFERENCES audience_briefs(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    geo JSONB NOT NULL DEFAULT '{}'::jsonb,
    ages JSONB,
    gender TEXT NOT NULL DEFAULT '',
    included_targeting JSONB NOT NULL DEFAULT '[]'::jsonb,
    excluded_targeting JSONB NOT NULL DEFAULT '[]'::jsonb,
    reach_estimate JSONB,
    confidence DOUBLE PRECISION,
    confidence_level TEXT NOT NULL DEFAULT 'experimental',
    basis JSONB NOT NULL DEFAULT '[]'::jsonb,
    explanation TEXT NOT NULL DEFAULT '',
    quality_score JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE offers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    segment_id UUID NOT NULL REFERENCES audience_segments(id) ON DELETE CASCADE,
    awareness_level TEXT NOT NULL DEFAULT '',
    problem TEXT NOT NULL DEFAULT '',
    motivation TEXT NOT NULL DEFAULT '',
    objection TEXT NOT NULL DEFAULT '',
    offer TEXT NOT NULL DEFAULT '',
    hook TEXT NOT NULL DEFAULT '',
    cta TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE creatives (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    segment_id UUID NOT NULL REFERENCES audience_segments(id) ON DELETE CASCADE,
    angle TEXT NOT NULL DEFAULT '',
    hook TEXT NOT NULL DEFAULT '',
    visual_idea TEXT NOT NULL DEFAULT '',
    copy_angle TEXT NOT NULL DEFAULT '',
    cta TEXT NOT NULL DEFAULT '',
    format TEXT NOT NULL DEFAULT 'reel',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE experiments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    brief_id UUID NOT NULL REFERENCES audience_briefs(id) ON DELETE CASCADE,
    hypothesis TEXT NOT NULL,
    audience TEXT NOT NULL DEFAULT '',
    variable TEXT NOT NULL DEFAULT '',
    variants JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    brief_id UUID REFERENCES audience_briefs(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    status TEXT NOT NULL,
    stage TEXT NOT NULL,
    error TEXT,
    payload JSONB,
    result JSONB,
    attempts INT NOT NULL DEFAULT 0,
    run_after TIMESTAMPTZ NOT NULL DEFAULT now(),
    locked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE targeting_cache (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    query TEXT NOT NULL,
    country TEXT NOT NULL DEFAULT '',
    kind TEXT NOT NULL,
    result JSONB NOT NULL,
    api_version TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (query, country, kind, api_version)
);

CREATE TABLE audit_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    entity_type TEXT NOT NULL,
    entity_id UUID,
    value TEXT NOT NULL DEFAULT '',
    source TEXT NOT NULL,
    meta_id TEXT,
    generated_from TEXT NOT NULL DEFAULT '',
    payload JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_projects_user ON projects(user_id);
CREATE INDEX idx_briefs_project ON audience_briefs(project_id);
CREATE INDEX idx_segments_brief ON audience_segments(brief_id);
CREATE INDEX idx_entities_brief ON targeting_entities(brief_id);
CREATE INDEX idx_strategies_brief ON audience_strategies(brief_id);
CREATE INDEX idx_jobs_status ON jobs(status, run_after);
CREATE INDEX idx_audit_project ON audit_events(project_id, created_at DESC);
