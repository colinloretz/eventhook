CREATE TABLE deliveries (
  id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  event_id             UUID NOT NULL REFERENCES events(id),
  endpoint_id          UUID NOT NULL REFERENCES endpoints(id),
  status               TEXT NOT NULL DEFAULT 'pending'
                       CHECK (status IN ('pending', 'delivered', 'failed', 'retrying')),
  attempt_count        INT DEFAULT 0,
  next_attempt         TIMESTAMPTZ DEFAULT NOW(),
  last_response_status INT,
  last_response_body   TEXT,
  last_attempted_at    TIMESTAMPTZ,
  delivered_at         TIMESTAMPTZ,
  created_at           TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_deliveries_pending ON deliveries(next_attempt)
  WHERE status IN ('pending', 'retrying');
CREATE INDEX idx_deliveries_event ON deliveries(event_id, created_at DESC);
CREATE INDEX idx_deliveries_status ON deliveries(status, created_at DESC);
