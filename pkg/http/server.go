package http

import (
	"context"
	"github.com/TicketsBot/logarchiver/internal"
	"github.com/TicketsBot/logarchiver/pkg/config"
	"github.com/TicketsBot/logarchiver/pkg/repository"
	"github.com/TicketsBot/logarchiver/pkg/repository/model"
	"github.com/TicketsBot/logarchiver/pkg/s3client"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"time"
)

type Server struct {
	Logger      *zap.Logger
	Config      config.Config
	RemoveQueue internal.RemoveQueue
	router      *gin.Engine
	store       repository.Store
	s3Clients   *s3client.ShardedClientManager
}

func NewServer(logger *zap.Logger, config config.Config, store repository.Store, clientManager *s3client.ShardedClientManager) *Server {
	return &Server{
		Logger:      logger,
		Config:      config,
		RemoveQueue: internal.NewRemoveQueue(logger),
		router:      gin.New(),
		store:       store,
		s3Clients:   clientManager,
	}
}

func (s *Server) RegisterRoutes() {
	s.router.Use(ginzap.Ginzap(s.Logger, time.RFC3339, true))
	s.router.Use(ginzap.RecoveryWithZap(s.Logger, true))

	s.router.GET("/", s.ticketGetHandler)
	s.router.POST("/", s.ticketUploadHandler)
	s.router.DELETE("/", s.ticketDeleteHandler)

	s.router.GET("/guild/status/:id", s.purgeStatusHandler)
	s.router.DELETE("/guild/:id", s.purgeGuildHandler)

	adminGroup := s.router.Group("/admin", s.middlewareAuthAdmin)
	{
		adminGroup.GET("/buckets", s.adminListBuckets)
		adminGroup.POST("/buckets", s.adminCreateBucket)
		adminGroup.PATCH("/buckets/active", s.adminSetActiveBucket)
	}
}

func (s *Server) Start() {
	if err := s.router.Run(s.Config.Address); err != nil {
		panic(err)
	}
}

func (s *Server) getActiveClient(ctx context.Context) (*s3client.S3Client, model.Bucket, error) {
	var bucket model.Bucket

	if err := s.store.Tx(ctx, func(r repository.Repositories) (err error) {
		bucket, err = r.Buckets().GetActiveBucket(ctx)
		return
	}); err != nil {
		return nil, model.Bucket{}, err
	}

	client, err := s.s3Clients.Get(bucket.Id)
	if err != nil {
		return nil, model.Bucket{}, err
	}

	return client, bucket, nil
}

func (s *Server) getClientForObject(ctx context.Context, guild uint64, ticket int) (*s3client.S3Client, bool, error) {
	bucketId := uuid.Nil
	if err := s.store.Tx(ctx, func(r repository.Repositories) error {
		object, ok, err := r.Objects().GetObject(ctx, guild, ticket)
		if err != nil {
			return err
		}

		if !ok {
			// TODO: Return status 404
			bucketId = s.Config.DefaultBucketId
		} else {
			bucketId = object.BucketId
		}

		return nil
	}); err != nil {
		return nil, false, err
	}

	if bucketId == uuid.Nil {
		return nil, false, nil
	}

	client, err := s.s3Clients.Get(bucketId)
	if err != nil {
		return nil, false, err
	}

	return client, true, nil
}
