package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func (s *PostgresStore) CreateDelivery(ctx context.Context, d *Delivery) error {
	return s.pool.QueryRow(ctx,
		`INSERT INTO deliveries (event_id, endpoint_id, status, next_attempt)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at`,
		d.EventID, d.EndpointID, d.Status, d.NextAttempt,
	).Scan(&d.ID, &d.CreatedAt)
}

func (s *PostgresStore) GetDelivery(ctx context.Context, id uuid.UUID) (*Delivery, error) {
	return scanDelivery(s.pool.QueryRow(ctx,
		`SELECT id, event_id, endpoint_id, status, attempt_count, next_attempt,
		        last_response_status, last_response_body, last_attempted_at, delivered_at, created_at
		 FROM deliveries WHERE id = $1`, id,
	))
}

func (s *PostgresStore) ListDeliveries(ctx context.Context, f ListDeliveriesFilter) ([]*Delivery, error) {
	where := []string{"1=1"}
	args := []any{}
	i := 1

	if f.EventID != nil {
		where = append(where, fmt.Sprintf("event_id = $%d", i))
		args = append(args, *f.EventID)
		i++
	}
	if f.EndpointID != nil {
		where = append(where, fmt.Sprintf("endpoint_id = $%d", i))
		args = append(args, *f.EndpointID)
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
		`SELECT id, event_id, endpoint_id, status, attempt_count, next_attempt,
		        last_response_status, last_response_body, last_attempted_at, delivered_at, created_at
		 FROM deliveries WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		strings.Join(where, " AND "), i, i+1,
	)
	args = append(args, limit, f.Offset)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deliveries []*Delivery
	for rows.Next() {
		d, err := scanDelivery(rows)
		if err != nil {
			return nil, err
		}
		deliveries = append(deliveries, d)
	}
	return deliveries, rows.Err()
}

func (s *PostgresStore) UpdateDelivery(ctx context.Context, d *Delivery) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE deliveries
		 SET status=$1, attempt_count=$2, next_attempt=$3,
		     last_response_status=$4, last_response_body=$5,
		     last_attempted_at=$6, delivered_at=$7
		 WHERE id=$8`,
		d.Status, d.AttemptCount, d.NextAttempt,
		d.LastResponseStatus, d.LastResponseBody,
		d.LastAttemptedAt, d.DeliveredAt, d.ID,
	)
	return err
}

// PollPendingDeliveries uses SKIP LOCKED for safe concurrent worker polling.
func (s *PostgresStore) PollPendingDeliveries(ctx context.Context, limit int) ([]*Delivery, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, event_id, endpoint_id, status, attempt_count, next_attempt,
		        last_response_status, last_response_body, last_attempted_at, delivered_at, created_at
		 FROM deliveries
		 WHERE status IN ('pending', 'retrying')
		   AND next_attempt <= NOW()
		 ORDER BY next_attempt ASC
		 LIMIT $1
		 FOR UPDATE SKIP LOCKED`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deliveries []*Delivery
	for rows.Next() {
		d, err := scanDelivery(rows)
		if err != nil {
			return nil, err
		}
		deliveries = append(deliveries, d)
	}
	return deliveries, rows.Err()
}

func (s *PostgresStore) CreateDeliveryAttempt(ctx context.Context, a *DeliveryAttempt) error {
	reqHeaders, _ := json.Marshal(a.RequestHeaders)
	respHeaders, _ := json.Marshal(a.ResponseHeaders)
	return s.pool.QueryRow(ctx,
		`INSERT INTO delivery_attempts
		 (delivery_id, attempt, status, request_headers, request_body,
		  response_status, response_headers, response_body, latency_ms)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id, attempted_at`,
		a.DeliveryID, a.Attempt, a.Status, reqHeaders, a.RequestBody,
		a.ResponseStatus, respHeaders, a.ResponseBody, a.LatencyMS,
	).Scan(&a.ID, &a.AttemptedAt)
}

func (s *PostgresStore) ListDeliveryAttempts(ctx context.Context, deliveryID uuid.UUID) ([]*DeliveryAttempt, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, delivery_id, attempt, status, request_headers, request_body,
		        response_status, response_headers, response_body, latency_ms, attempted_at
		 FROM delivery_attempts WHERE delivery_id = $1 ORDER BY attempt ASC`,
		deliveryID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attempts []*DeliveryAttempt
	for rows.Next() {
		a, err := scanDeliveryAttempt(rows)
		if err != nil {
			return nil, err
		}
		attempts = append(attempts, a)
	}
	return attempts, rows.Err()
}

func scanDelivery(row scanner) (*Delivery, error) {
	var d Delivery
	err := row.Scan(
		&d.ID, &d.EventID, &d.EndpointID, &d.Status, &d.AttemptCount, &d.NextAttempt,
		&d.LastResponseStatus, &d.LastResponseBody, &d.LastAttemptedAt, &d.DeliveredAt, &d.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan delivery: %w", err)
	}
	return &d, nil
}

func scanDeliveryAttempt(row scanner) (*DeliveryAttempt, error) {
	var a DeliveryAttempt
	var reqHeaders, respHeaders []byte
	err := row.Scan(
		&a.ID, &a.DeliveryID, &a.Attempt, &a.Status, &reqHeaders, &a.RequestBody,
		&a.ResponseStatus, &respHeaders, &a.ResponseBody, &a.LatencyMS, &a.AttemptedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan delivery attempt: %w", err)
	}
	if reqHeaders != nil {
		_ = json.Unmarshal(reqHeaders, &a.RequestHeaders)
	}
	if respHeaders != nil {
		_ = json.Unmarshal(respHeaders, &a.ResponseHeaders)
	}
	return &a, nil
}
