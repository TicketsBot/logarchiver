package repository

import "github.com/jackc/pgx/v5"

type PostgresRepositories struct {
	tx pgx.Tx
}

func newPostgresRepositories(tx pgx.Tx) *PostgresRepositories {
	return &PostgresRepositories{
		tx: tx,
	}
}

var _ Repositories = (*PostgresRepositories)(nil)

func (p *PostgresRepositories) Buckets() BucketRepository {
	return newPostgresBucketRepository(p.tx)
}

func (p *PostgresRepositories) Objects() ObjectRepository {
	return newPostgresObjectRepository(p.tx)
}
