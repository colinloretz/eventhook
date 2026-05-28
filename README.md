# EventHook

Stripe-quality webhook observability for every event in your app.

Drop in a library, mount a dashboard, and immediately see every inbound and outbound webhook flowing through your application — with full payloads, delivery history, and one-click replay.

```
EventHook.emit('payment.completed', { order_id: 123, amount: 9900 })
```

---

## What it is

EventHook is a self-hosted webhook infrastructure runtime. It handles:

- **Inbound webhooks** — receive, verify signatures, and record every Stripe, GitHub, or custom webhook that hits your app
- **Outbound delivery** — fan out events to your registered endpoints with automatic retry and exponential backoff
- **Full observability** — every payload, every delivery attempt, every response body, all queryable in a live dashboard

The runtime is a single Go binary. Language libraries are thin HTTP clients. The dashboard is embedded in the binary — nothing to deploy separately.

---

## Quickstart

**Requires:** Docker

```bash
git clone https://github.com/eventhook/eventhook
cd eventhook
docker compose up
```

Open **http://localhost:7676/dashboard**

Send your first event:

```bash
curl -X POST http://localhost:7676/api/v1/events \
  -H "Authorization: Bearer dev-api-key" \
  -H "Content-Type: application/json" \
  -d '{"event_type": "payment.completed", "payload": {"order_id": 123, "amount": 9900}}'
```

Refresh the dashboard. The event appears instantly.

---

## Dashboard

Five views, all live-updating:

| View | URL | What you see |
|------|-----|--------------|
| Event Stream | `/dashboard/` | Every event, color-coded by status, polling every 2s |
| Event Detail | `/dashboard/events/:id` | Full payload, headers, delivery status, replay button |
| Deliveries | `/dashboard/deliveries` | All outbound attempts, filterable by status |
| Delivery Detail | `/dashboard/deliveries/:id` | Every attempt with request + response bodies side-by-side |
| Endpoints | `/dashboard/endpoints` | Create, edit, enable/disable outbound destinations |

---

## Architecture

```
Your App
  │
  │  POST /api/v1/events          (emit outbound events)
  │  POST /api/v1/in/:source_slug (receive inbound webhooks)
  ▼
EventHook Runtime (Go)
  ├── API Layer (gin)
  ├── Worker Pool (SKIP LOCKED, configurable goroutine count)
  └── Dashboard (React, embedded via go:embed)
  │
  ├── Postgres  (event storage, delivery queue)
  └── Redis     (future: pub/sub for SSE)
```

The worker pool polls Postgres using `SELECT ... FOR UPDATE SKIP LOCKED`, which means you can run multiple runtime instances horizontally without duplicate deliveries.

---

## Configuration

All config via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `EVENTHOOK_DATABASE_URL` | `postgres://eventhook:eventhook@localhost:5432/eventhook?sslmode=disable` | Postgres connection string |
| `EVENTHOOK_REDIS_URL` | `redis://localhost:6379` | Redis connection string |
| `EVENTHOOK_API_KEY` | `dev-api-key` | Bearer token for API authentication |
| `EVENTHOOK_PORT` | `7676` | HTTP port |
| `EVENTHOOK_WORKER_COUNT` | `10` | Delivery goroutine count |
| `EVENTHOOK_MAX_PAYLOAD_KB` | `1024` | Max inbound payload size |
| `EVENTHOOK_ENV` | `development` | `development` or `production` |

---

## API

All endpoints require `Authorization: Bearer <api_key>` except `/health`.

### Events

```
POST   /api/v1/events              Ingest an event
GET    /api/v1/events              List events (filter: source_id, event_type, status)
GET    /api/v1/events/:id          Event detail
POST   /api/v1/events/:id/replay   Replay event (creates new event + deliveries)
```

### Inbound Webhooks

```
POST   /api/v1/in/:source_slug     Receive inbound webhook (no auth — uses source secret)
```

### Deliveries

```
GET    /api/v1/deliveries          List deliveries (filter: event_id, endpoint_id, status)
GET    /api/v1/deliveries/:id      Delivery detail + all attempts
POST   /api/v1/deliveries/:id/retry  Manual retry
```

### Endpoints

```
GET    /api/v1/endpoints           List endpoints
POST   /api/v1/endpoints           Create endpoint
PUT    /api/v1/endpoints/:id       Update endpoint
DELETE /api/v1/endpoints/:id       Delete endpoint
```

### Sources

```
GET    /api/v1/sources             List sources
POST   /api/v1/sources             Create source
PUT    /api/v1/sources/:id         Update source
DELETE /api/v1/sources/:id         Delete source
```

### Ingest payload

```json
POST /api/v1/events
{
  "event_type": "payment.completed",
  "payload": { "order_id": 123, "amount": 9900 },
  "source_id": "uuid-optional",
  "idempotency_key": "optional-unique-key"
}
```

---

## Retry Schedule

Failed deliveries are retried automatically:

| Attempt | Delay |
|---------|-------|
| 1 | 30 seconds |
| 2 | 5 minutes |
| 3 | 30 minutes |
| 4 | 2 hours |
| 5 | 5 hours |
| 6+ | Marked failed |

---

## Building from source

```bash
# Build dashboard + Go binary
make build

# Output: bin/eventhook
./bin/eventhook
```

**Requires:** Go 1.22+, Node 18+

The Makefile builds the React app, copies the dist into `assets/dashboard/`, and compiles the Go binary with assets embedded. The result is a single self-contained binary.

---

## Development

```bash
# Zero-config dev mode — auto-starts postgres + redis via docker, runs migrations,
# serves the runtime, and tails events live in the terminal
make build
./bin/eventhook dev
```

```bash
# Dashboard hot-reload (proxies API calls to :7676)
cd dashboard && npm run dev
# http://localhost:5173/dashboard/
```

---

## Roadmap

- [x] Go runtime — API, worker pool, retry
- [x] React dashboard — event stream, delivery detail, endpoint management
- [x] Rails gem — `EventHook.emit`, engine mount, Stripe/GitHub/Shopify verification, dashboard proxy
- [x] `eventhook dev` CLI — auto-starts postgres/redis, live event tail, cross-platform release builds
- [ ] Laravel package
- [ ] Node/TypeScript SDK
- [ ] Managed cloud (eventhook.dev)

---

## License

MIT
