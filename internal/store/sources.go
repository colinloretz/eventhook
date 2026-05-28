package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

func (s *PostgresStore) CreateSource(ctx context.Context, src *Source) error {
	meta, _ := json.Marshal(src.Metadata)
	return s.pool.QueryRow(ctx,
		`INSERT INTO sources (name, slug, secret, source_type, metadata)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at`,
		src.Name, src.Slug, src.Secret, src.SourceType, meta,
	).Scan(&src.ID, &src.CreatedAt)
}

func (s *PostgresStore) GetSource(ctx context.Context, id uuid.UUID) (*Source, error) {
	return scanSource(s.pool.QueryRow(ctx,
		`SELECT id, name, slug, secret, source_type, metadata, created_at FROM sources WHERE id = $1`, id,
	))
}

func (s *PostgresStore) GetSourceBySlug(ctx context.Context, slug string) (*Source, error) {
	return scanSource(s.pool.QueryRow(ctx,
		`SELECT id, name, slug, secret, source_type, metadata, created_at FROM sources WHERE slug = $1`, slug,
	))
}

func (s *PostgresStore) ListSources(ctx context.Context) ([]*Source, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, name, slug, secret, source_type, metadata, created_at FROM sources ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []*Source
	for rows.Next() {
		src, err := scanSource(rows)
		if err != nil {
			return nil, err
		}
		sources = append(sources, src)
	}
	return sources, rows.Err()
}

func (s *PostgresStore) UpdateSource(ctx context.Context, src *Source) error {
	meta, _ := json.Marshal(src.Metadata)
	_, err := s.pool.Exec(ctx,
		`UPDATE sources SET name=$1, slug=$2, secret=$3, source_type=$4, metadata=$5 WHERE id=$6`,
		src.Name, src.Slug, src.Secret, src.SourceType, meta, src.ID,
	)
	return err
}

func (s *PostgresStore) DeleteSource(ctx context.Context, id uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM sources WHERE id=$1`, id)
	return err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanSource(row scanner) (*Source, error) {
	var src Source
	var metaJSON []byte
	err := row.Scan(&src.ID, &src.Name, &src.Slug, &src.Secret, &src.SourceType, &metaJSON, &src.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("scan source: %w", err)
	}
	if metaJSON != nil {
		_ = json.Unmarshal(metaJSON, &src.Metadata)
	}
	return &src, nil
}
