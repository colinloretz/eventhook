CREATE TABLE endpoints (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  url         TEXT NOT NULL,
  description TEXT,
  secret      TEXT NOT NULL,
  enabled     BOOLEAN DEFAULT true,
  event_types TEXT[],
  metadata    JSONB DEFAULT '{}',
  created_at  TIMESTAMPTZ DEFAULT NOW()
);
