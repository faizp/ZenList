package repo

import (
	"context"

	"github.com/faizp/zenlist/backend/go-graphql/internal/db/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store holds query dependencies and provides transaction helpers.
type Store struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool, q: sqlc.New(pool)}
}

func (s *Store) Queries() *sqlc.Queries {
	return s.q
}

func (s *Store) Pool() *pgxpool.Pool {
	return s.pool
}

func (s *Store) WithTx(ctx context.Context, fn func(*sqlc.Queries) error) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := sqlc.New(tx)
	if err := fn(qtx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
