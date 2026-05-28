export interface Source {
  id: string
  name: string
  slug: string
  source_type: 'inbound' | 'outbound' | 'internal'
  metadata: Record<string, unknown>
  created_at: string
}

export interface Event {
  id: string
  source_id: string | null
  event_type: string
  payload: Record<string, unknown>
  headers: Record<string, unknown>
  idempotency_key: string | null
  status: 'pending' | 'delivered' | 'failed' | 'ignored'
  payload_truncated: boolean
  received_at: string
  created_at: string
}

export interface Endpoint {
  id: string
  url: string
  description: string | null
  enabled: boolean
  event_types: string[]
  metadata: Record<string, unknown>
  created_at: string
}

export interface Delivery {
  id: string
  event_id: string
  endpoint_id: string
  status: 'pending' | 'delivered' | 'failed' | 'retrying'
  attempt_count: number
  next_attempt: string
  last_response_status: number | null
  last_response_body: string | null
  last_attempted_at: string | null
  delivered_at: string | null
  created_at: string
}

export interface DeliveryAttempt {
  id: string
  delivery_id: string
  attempt: number
  status: 'success' | 'failure' | 'timeout'
  request_headers: Record<string, unknown> | null
  request_body: string | null
  response_status: number | null
  response_headers: Record<string, unknown> | null
  response_body: string | null
  latency_ms: number | null
  attempted_at: string
}
