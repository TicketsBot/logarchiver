package repository

import (
	"context"
	_ "embed"
	"errors"
	"github.com/TicketsBot/logarchiver/pkg/repository/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type PostgresBucketRepository struct {
	tx pgx.Tx
}

var _ BucketRepository = (*PostgresBucketRepository)(nil)

var (
	//go:embed sql/buckets/list.sql
	queryListBuckets string

	//go:embed sql/buckets/get_active.sql
	queryGetActiveBucket string

	//go:embed sql/buckets/set_active_remove_old.sql
	querySetActiveBucketRemoveOld string

	//go:embed sql/buckets/set_active_set_new.sql
	querySetActiveBucketSetNew string

	//go:embed sql/buckets/create.sql
	queryCreateBucket string

	//go:embed sql/buckets/delete.sql
	queryDeleteBucket string
)

func newPostgresBucketRepository(tx pgx.Tx) *PostgresBucketRepository {
	return &PostgresBucketRepository{
		tx: tx,
	}
}

func (p *PostgresBucketRepository) ListBuckets(ctx context.Context) ([]model.Bucket, error) {
	res, err := p.tx.Query(ctx, queryListBuckets)
	if err != nil {
		return nil, err
	}

	defer res.Close()

	buckets := make([]model.Bucket, 0)
	for res.Next() {
		var bucket model.Bucket
		if err := res.Scan(&bucket.Id, &bucket.EndpointUrl, &bucket.Name, &bucket.Active); err != nil {
			return nil, err
		}

		buckets = append(buckets, bucket)
	}

	return buckets, nil
}

func (p *PostgresBucketRepository) GetActiveBucket(ctx context.Context) (model.Bucket, error) {
	var bucket model.Bucket
	if err := p.tx.QueryRow(ctx, queryGetActiveBucket).Scan(&bucket.Id, &bucket.EndpointUrl, &bucket.Name, &bucket.Active); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Bucket{}, ErrNoActiveBucket
		} else {
			return model.Bucket{}, err
		}
	}

	return bucket, nil
}

func (p *PostgresBucketRepository) SetActiveBucket(ctx context.Context, id uuid.UUID) error {
	if _, err := p.tx.Exec(ctx, querySetActiveBucketRemoveOld); err != nil {
		return err
	}

	if _, err := p.tx.Exec(ctx, querySetActiveBucketSetNew, id); err != nil {
		return err
	}

	return nil
}

func (p *PostgresBucketRepository) CreateBucket(ctx context.Context, endpointUrl, name string) (uuid.UUID, error) {
	id := uuid.New()

	if _, err := p.tx.Exec(ctx, queryCreateBucket, id, endpointUrl, name); err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (p *PostgresBucketRepository) DeleteBucket(ctx context.Context, id uuid.UUID) error {
	if _, err := p.tx.Exec(ctx, queryDeleteBucket, id); err != nil {
		return err
	}

	return nil
}
