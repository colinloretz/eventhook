package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Source struct {
	ID         uuid.UUID      `json:"id"`
	Name       string         `json:"name"`
	Slug       string         `json:"slug"`
	Secret     *string        `json:"secret,omitempty"`
	SourceType string         `json:"source_type"`
	Metadata   map[string]any `json:"metadata"`
	CreatedAt  time.Time      `json:"created_at"`
}

type Event struct {
	ID                uuid.UUID      `json:"id"`
	SourceID          *uuid.UUID     `json:"source_id,omitempty"`
	EventType         string         `json:"event_type"`
	Payload           map[string]any `json:"payload"`
	Headers           map[string]any `json:"headers"`
	IdempotencyKey    *string        `json:"idempotency_key,omitempty"`
	Status            string         `json:"status"`
	PayloadTruncated  bool           `json:"payload_truncated"`
	PayloadStorageURL *string        `json:"payload_storage_url,omitempty"`
	ReceivedAt        time.Time      `json:"received_at"`
	CreatedAt         time.Time      `json:"created_at"`
}

type Endpoint struct {
	ID          uuid.UUID      `json:"id"`
	URL         string         `json:"url"`
	Description *string        `json:"description,omitempty"`
	Secret      string         `json:"-"`
	Enabled     bool           `json:"enabled"`
	EventTypes  []string       `json:"event_types"`
	Metadata    map[string]any `json:"metadata"`
	CreatedAt   time.Time      `json:"created_at"`
}

type Delivery struct {
	ID                 uuid.UUID  `json:"id"`
	EventID            uuid.UUID  `json:"event_id"`
	EndpointID         uuid.UUID  `json:"endpoint_id"`
	Status             string     `json:"status"`
	AttemptCount       int        `json:"attempt_count"`
	NextAttempt        time.Time  `json:"next_attempt"`
	LastResponseStatus *int       `json:"last_response_status,omitempty"`
	LastResponseBody   *string    `json:"last_response_body,omitempty"`
	LastAttemptedAt    *time.Time `json:"last_attempted_at,omitempty"`
	DeliveredAt        *time.Time `json:"delivered_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
}

type DeliveryAttempt struct {
	ID              uuid.UUID      `json:"id"`
	DeliveryID      uuid.UUID      `json:"delivery_id"`
	Attempt         int            `json:"attempt"`
	Status          string         `json:"status"`
	RequestHeaders  map[string]any `json:"request_headers,omitempty"`
	RequestBody     *string        `json:"request_body,omitempty"`
	ResponseStatus  *int           `json:"response_status,omitempty"`
	ResponseHeaders map[string]any `json:"response_headers,omitempty"`
	ResponseBody    *string        `json:"response_body,omitempty"`
	LatencyMS       *int           `json:"latency_ms,omitempty"`
	AttemptedAt     time.Time      `json:"attempted_at"`
}

type ListEventsFilter struct {
	SourceID  *uuid.UUID
	EventType *string
	Status    *string
	Limit     int
	Offset    int
}

type ListDeliveriesFilter struct {
	EventID    *uuid.UUID
	EndpointID *uuid.UUID
	Status     *string
	Limit      int
	Offset     int
}

type Store interface {
	// Sources
	CreateSource(ctx context.Context, s *Source) error
	GetSource(ctx context.Context, id uuid.UUID) (*Source, error)
	GetSourceBySlug(ctx context.Context, slug string) (*Source, error)
	ListSources(ctx context.Context) ([]*Source, error)
	UpdateSource(ctx context.Context, s *Source) error
	DeleteSource(ctx context.Context, id uuid.UUID) error

	// Events
	CreateEvent(ctx context.Context, e *Event) error
	GetEvent(ctx context.Context, id uuid.UUID) (*Event, error)
	ListEvents(ctx context.Context, f ListEventsFilter) ([]*Event, error)
	UpdateEventStatus(ctx context.Context, id uuid.UUID, status string) error

	// Endpoints
	CreateEndpoint(ctx context.Context, ep *Endpoint) error
	GetEndpoint(ctx context.Context, id uuid.UUID) (*Endpoint, error)
	ListEndpoints(ctx context.Context) ([]*Endpoint, error)
	UpdateEndpoint(ctx context.Context, ep *Endpoint) error
	DeleteEndpoint(ctx context.Context, id uuid.UUID) error

	// Deliveries
	CreateDelivery(ctx context.Context, d *Delivery) error
	GetDelivery(ctx context.Context, id uuid.UUID) (*Delivery, error)
	ListDeliveries(ctx context.Context, f ListDeliveriesFilter) ([]*Delivery, error)
	UpdateDelivery(ctx context.Context, d *Delivery) error
	PollPendingDeliveries(ctx context.Context, limit int) ([]*Delivery, error)

	// Delivery Attempts
	CreateDeliveryAttempt(ctx context.Context, a *DeliveryAttempt) error
	ListDeliveryAttempts(ctx context.Context, deliveryID uuid.UUID) ([]*DeliveryAttempt, error)

	Close()
}
