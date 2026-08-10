package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBTX — общий набор методов для pool и транзакции.
type DBTX interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
}

type Repository struct {
	db DBTX
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{db: pool}
}

// Tx выполняет fn внутри одной транзакции (откат при ошибке).
func (r *Repository) Tx(ctx context.Context, fn func(*Repository) error) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	if err := fn(&Repository{db: tx}); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}
