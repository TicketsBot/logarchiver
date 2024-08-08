package s3client

import (
	"context"
	"errors"
	"github.com/TicketsBot/logarchiver/pkg/config"
	"github.com/TicketsBot/logarchiver/pkg/repository"
	"github.com/TicketsBot/logarchiver/pkg/repository/model"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"sync"
)

type ShardedClientManager struct {
	config  config.Config
	store   repository.Store
	clients map[uuid.UUID]*S3Client
	mu      sync.RWMutex
}

var ErrClientNotFound = errors.New("client not found")

func NewShardedClientManager(config config.Config, store repository.Store) *ShardedClientManager {
	return &ShardedClientManager{
		config:  config,
		store:   store,
		clients: make(map[uuid.UUID]*S3Client),
	}
}

func (s *ShardedClientManager) Load(ctx context.Context) error {
	var buckets []model.Bucket

	if err := s.store.Tx(ctx, func(r repository.Repositories) error {
		tmp, err := r.Buckets().ListBuckets(ctx)
		if err != nil {
			return err
		}

		buckets = tmp

		return nil
	}); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.clients = make(map[uuid.UUID]*S3Client)

	for _, bucket := range buckets {
		m, err := minio.New(bucket.EndpointUrl, &minio.Options{
			Creds:  credentials.NewStaticV4(s.config.AccessKey, s.config.SecretKey, ""),
			Secure: true,
		})

		if err != nil {
			return err
		}

		s.clients[bucket.Id] = NewS3Client(m, bucket.EndpointUrl)
	}

	return nil
}

func (s *ShardedClientManager) Get(bucketId uuid.UUID) (*S3Client, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, ok := s.clients[bucketId]
	if !ok {
		return nil, ErrClientNotFound
	}

	return client, nil
}

func (s *ShardedClientManager) GetAll() []*S3Client {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clients := make([]*S3Client, 0, len(s.clients))
	for _, client := range s.clients {
		clients = append(clients, client)
	}

	return clients
}
