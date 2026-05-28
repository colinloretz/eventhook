package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func (s *PostgresStore) CreateEvent(ctx context.Context, e *Event) error {
	payload, _ := json.Marshal(e.Payload)
	headers, _ := json.Marshal(e.Headers)
	return s.pool.QueryRow(ctx,
		`INSERT INTO events (source_id, event_type, payload, headers, idempotency_key, status, payload_truncated)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, received_at, created_at`,
		e.SourceID, e.EventType, payload, headers, e.IdempotencyKey, e.Status, e.PayloadTruncated,
	).Scan(&e.ID, &e.ReceivedAt, &e.CreatedAt)
}

func (s *PostgresStore) GetEvent(ctx context.Context, id uuid.UUID) (*Event, error) {
	return scanEvent(s.pool.QueryRow(ctx,
		`SELECT id, source_id, event_type, payload, headers, idempotency_key, status,
		        payload_truncated, payload_storage_url, received_at, created_at
		 FROM events WHERE id = $1`, id,
	))
}

func (s *PostgresStore) ListEvents(ctx context.Context, f ListEventsFilter) ([]*Event, error) {
	where := []string{"1=1"}
	args := []any{}
	i := 1

	if f.SourceID != nil {
		where = append(where, fmt.Sprintf("source_id = $%d", i))
		args = append(args, *f.SourceID)
		i++
	}
	if f.EventType != nil {
		where = append(where, fmt.Sprintf("event_type = $%d", i))
		args = append(args, *f.EventType)
		i++
	}
	if f.Status != nil {
		where = append(where, fmt.Sprintf("status = $%d", i))
		args = append(args, *f.Status)
		i++
	}

	limit := f.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	query := fmt.Sprintf(
		`SELECT id, source_id, event_type, payload, headers, idempotency_key, status,
		        payload_truncated, payload_storage_url, received_at, created_at
		 FROM events WHERE %s ORDER BY received_at DESC LIMIT $%d OFFSET $%d`,
		strings.Join(where, " AND "), i, i+1,
	)
	args = append(args, limit, f.Offset)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*Event
	for rows.Next() {
		ev, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, ev)
	}
	return events, rows.Err()
}

func (s *PostgresStore) UpdateEventStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := s.pool.Exec(ctx, `UPDATE events SET status=$1 WHERE id=$2`, status, id)
	return err
}

func scanEvent(row scanner) (*Event, error) {
	var e Event
	var payloadJSON, headersJSON []byte
	err := row.Scan(
		&e.ID, &e.SourceID, &e.EventType, &payloadJSON, &headersJSON,
		&e.IdempotencyKey, &e.Status, &e.PayloadTruncated, &e.PayloadStorageURL,
		&e.ReceivedAt, &e.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan event: %w", err)
	}
	if payloadJSON != nil {
		_ = json.Unmarshal(payloadJSON, &e.Payload)
	}
	if headersJSON != nil {
		_ = json.Unmarshal(headersJSON, &e.Headers)
	}
	return &e, nil
}
