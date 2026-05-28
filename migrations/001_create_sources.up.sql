CREATE TABLE sources (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name        TEXT NOT NULL,
  slug        TEXT NOT NULL UNIQUE,
  secret      TEXT,
  source_type TEXT NOT NULL CHECK (source_type IN ('inbound', 'outbound', 'internal')),
  metadata    JSONB DEFAULT '{}',
  created_at  TIMESTAMPTZ DEFAULT NOW()
);
