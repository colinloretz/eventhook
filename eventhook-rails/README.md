# eventhook-rails


Drop-in webhook observability for Rails. See every inbound and outbound webhook in a live dashboard — with full payloads, delivery history, and one-click replay.

## Installation

Add to your Gemfile:

```ruby
gem 'eventhook'
```

Run the EventHook runtime (see [eventhook](../README.md)):

```bash
docker compose up  # from the eventhook repo root
```

## Setup

```ruby
# config/initializers/eventhook.rb
EventHook.configure do |config|
  config.runtime_url = ENV['EVENTHOOK_URL']  # default: http://localhost:7676
  config.api_key     = ENV['EVENTHOOK_KEY']  # default: dev-api-key

  config.sources do |s|
    s.add :stripe,  secret: ENV['STRIPE_WEBHOOK_SECRET']
    s.add :github,  secret: ENV['GITHUB_WEBHOOK_SECRET']
    s.add :shopify, secret: ENV['SHOPIFY_WEBHOOK_SECRET']
  end
end
```

```ruby
# config/routes.rb
Rails.application.routes.draw do
  mount EventHook::Engine, at: '/eventhook'
end
```

This mounts:

| Route | Description |
|-------|-------------|
| `POST /eventhook/in/:source` | Receive + verify inbound webhooks |
| `GET  /eventhook/dashboard`  | Live webhook dashboard (proxied from runtime) |
| `*    /eventhook/api/*`      | Runtime API proxy (used by dashboard) |

## Emitting events

```ruby
# From anywhere in your app
EventHook.emit('payment.completed', { order_id: order.id, amount: order.total })

# With idempotency key
EventHook.emit('user.created',
  { id: user.id, email: user.email },
  idempotency_key: "user-created-#{user.id}"
)
```

## Receiving inbound webhooks

Point your webhook provider at `/eventhook/in/:source`, where `:source` matches the slug you registered in the initializer.

**Stripe:** `https://yourapp.com/eventhook/in/stripe`
**GitHub:** `https://yourapp.com/eventhook/in/github`
**Shopify:** `https://yourapp.com/eventhook/in/shopify`

Signatures are verified automatically:

| Provider | Signature scheme |
|----------|-----------------|
| `stripe` | `Stripe-Signature` HMAC-SHA256 with timestamp tolerance |
| `github` | `X-Hub-Signature-256` HMAC-SHA256 |
| `shopify` | `X-Shopify-Hmac-SHA256` Base64 HMAC-SHA256 |
| custom | Generic HMAC-SHA256 with configurable header |

Invalid signatures return `401`. Verified events are forwarded to the runtime and appear in the dashboard immediately.

## Dashboard

Open `http://localhost:3000/eventhook/dashboard` to see all events flowing through your app.

The dashboard is reverse-proxied from the EventHook runtime — you get the same UI regardless of which language library you use.

## Custom / generic sources

```ruby
EventHook.configure do |config|
  config.sources do |s|
    # Uses X-Webhook-Signature: sha256=<hmac> by default
    s.add :my_provider, secret: ENV['MY_SECRET'], type: :generic
  end
end
```
