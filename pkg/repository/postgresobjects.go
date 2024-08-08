package repository

import (
	"context"
	_ "embed"
	"errors"
	"github.com/TicketsBot/logarchiver/pkg/repository/model"
	"github.com/jackc/pgx/v5"
)

type PostgresObjectRepository struct {
	tx pgx.Tx
}

var _ ObjectRepository = (*PostgresObjectRepository)(nil)

var (
	//go:embed sql/objects/get.sql
	queryGetObject string

	//go:embed sql/objects/create.sql
	queryCreateObject string

	//go:embed sql/objects/delete.sql
	queryDeleteObject string
)

func newPostgresObjectRepository(tx pgx.Tx) *PostgresObjectRepository {
	return &PostgresObjectRepository{
		tx: tx,
	}
}

func (p *PostgresObjectRepository) GetObject(ctx context.Context, guildId uint64, ticketId int) (model.Object, bool, error) {
	var object model.Object
	if err := p.tx.QueryRow(ctx, queryGetObject, guildId, ticketId).Scan(&object.GuildId, &object.TicketId, &object.BucketId); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Object{}, false, nil
		} else {
			return model.Object{}, false, err
		}
	}

	return object, true, nil
}

func (p *PostgresObjectRepository) CreateObject(ctx context.Context, object model.Object) error {
	if _, err := p.tx.Exec(ctx, queryCreateObject, object.GuildId, object.TicketId, object.BucketId); err != nil {
		return err
	}

	return nil
}

func (p *PostgresObjectRepository) DeleteObject(ctx context.Context, guildId uint64, ticketId int) error {
	if _, err := p.tx.Exec(ctx, queryDeleteObject, guildId, ticketId); err != nil {
		return err
	}

	return nil
}
