package repository

import (
	"context"
	"github.com/TicketsBot/logarchiver/pkg/repository/model"
	"github.com/google/uuid"
)

type Store interface {
	Tx(context.Context, func(Repositories) error) error
}

type Repositories interface {
	Buckets() BucketRepository
	Objects() ObjectRepository
}

type BucketRepository interface {
	ListBuckets(ctx context.Context) ([]model.Bucket, error)
	GetActiveBucket(ctx context.Context) (model.Bucket, error)
	SetActiveBucket(ctx context.Context, id uuid.UUID) error
	CreateBucket(ctx context.Context, endpointUrl, name string) (uuid.UUID, error)
	DeleteBucket(ctx context.Context, id uuid.UUID) error
}

type ObjectRepository interface {
	GetObject(ctx context.Context, guildId uint64, ticketId int) (model.Object, bool, error)
	CreateObject(ctx context.Context, object model.Object) error
	DeleteObject(ctx context.Context, guildId uint64, ticketId int) error
}
