package repository

import (
	"context"
	"github.com/TicketsBot/logarchiver/pkg/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	db *pgxpool.Pool
}

var _ Store = (*PostgresStore)(nil)

func NewPostgresRepository(db *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{
		db: db,
	}
}

func ConnectPostgres(ctx context.Context, config config.Config) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, config.DatabaseUri)
	if err != nil {
		return nil, err
	}

	return NewPostgresRepository(pool), nil
}

func (p *PostgresStore) Tx(ctx context.Context, f func(Repositories) error) error {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	if err := f(newPostgresRepositories(tx)); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
