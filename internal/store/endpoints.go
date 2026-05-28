package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

func (s *PostgresStore) CreateEndpoint(ctx context.Context, ep *Endpoint) error {
	meta, _ := json.Marshal(ep.Metadata)
	return s.pool.QueryRow(ctx,
		`INSERT INTO endpoints (url, description, secret, enabled, event_types, metadata)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, created_at`,
		ep.URL, ep.Description, ep.Secret, ep.Enabled, ep.EventTypes, meta,
	).Scan(&ep.ID, &ep.CreatedAt)
}

func (s *PostgresStore) GetEndpoint(ctx context.Context, id uuid.UUID) (*Endpoint, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT id, url, description, secret, enabled, event_types, metadata, created_at
		 FROM endpoints WHERE id = $1`, id,
	)
	return scanEndpoint(row)
}

func (s *PostgresStore) ListEndpoints(ctx context.Context) ([]*Endpoint, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, url, description, secret, enabled, event_types, metadata, created_at
		 FROM endpoints ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var endpoints []*Endpoint
	for rows.Next() {
		ep, err := scanEndpoint(rows)
		if err != nil {
			return nil, err
		}
		endpoints = append(endpoints, ep)
	}
	return endpoints, rows.Err()
}

func (s *PostgresStore) UpdateEndpoint(ctx context.Context, ep *Endpoint) error {
	meta, _ := json.Marshal(ep.Metadata)
	_, err := s.pool.Exec(ctx,
		`UPDATE endpoints SET url=$1, description=$2, secret=$3, enabled=$4, event_types=$5, metadata=$6 WHERE id=$7`,
		ep.URL, ep.Description, ep.Secret, ep.Enabled, ep.EventTypes, meta, ep.ID,
	)
	return err
}

func (s *PostgresStore) DeleteEndpoint(ctx context.Context, id uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM endpoints WHERE id=$1`, id)
	return err
}

func scanEndpoint(row scanner) (*Endpoint, error) {
	var ep Endpoint
	var metaJSON []byte
	err := row.Scan(
		&ep.ID, &ep.URL, &ep.Description, &ep.Secret, &ep.Enabled,
		&ep.EventTypes, &metaJSON, &ep.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan endpoint: %w", err)
	}
	if metaJSON != nil {
		_ = json.Unmarshal(metaJSON, &ep.Metadata)
	}
	return &ep, nil
}
