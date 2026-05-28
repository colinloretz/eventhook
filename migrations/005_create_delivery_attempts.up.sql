CREATE TABLE delivery_attempts (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  delivery_id      UUID NOT NULL REFERENCES deliveries(id),
  attempt          INT NOT NULL,
  status           TEXT NOT NULL CHECK (status IN ('success', 'failure', 'timeout')),
  request_headers  JSONB,
  request_body     TEXT,
  response_status  INT,
  response_headers JSONB,
  response_body    TEXT,
  latency_ms       INT,
  attempted_at     TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_delivery_attempts_delivery ON delivery_attempts(delivery_id, attempted_at DESC);
