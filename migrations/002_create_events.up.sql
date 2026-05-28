CREATE TABLE events (
  id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  source_id           UUID REFERENCES sources(id),
  event_type          TEXT NOT NULL,
  payload             JSONB NOT NULL,
  headers             JSONB DEFAULT '{}',
  idempotency_key     TEXT UNIQUE,
  status              TEXT NOT NULL DEFAULT 'pending'
                      CHECK (status IN ('pending', 'delivered', 'failed', 'ignored')),
  payload_truncated   BOOLEAN DEFAULT false,
  payload_storage_url TEXT,
  received_at         TIMESTAMPTZ DEFAULT NOW(),
  created_at          TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_events_received_at ON events(received_at DESC);
CREATE INDEX idx_events_source_type ON events(source_id, event_type, received_at DESC);
CREATE INDEX idx_events_idempotency ON events(idempotency_key) WHERE idempotency_key IS NOT NULL;
CREATE INDEX idx_events_status ON events(status, received_at DESC);
