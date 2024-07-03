package http

import (
	"github.com/TicketsBot/logarchiver/internal"
	"github.com/TicketsBot/logarchiver/pkg/config"
	"github.com/TicketsBot/logarchiver/pkg/s3client"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
	"time"
)

type Server struct {
	Logger      *zap.Logger
	Config      config.Config
	RemoveQueue internal.RemoveQueue
	router      *gin.Engine
	minio       *minio.Client
	s3          *s3client.S3Client
}

func NewServer(logger *zap.Logger, config config.Config, client *minio.Client) *Server {
	return &Server{
		Logger:      logger,
		Config:      config,
		RemoveQueue: internal.NewRemoveQueue(logger),
		router:      gin.New(),
		minio:       client,
		s3:          s3client.NewS3Client(client, config.Bucket),
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
}

func (s *Server) Start() {
	if err := s.router.Run(s.Config.Address); err != nil {
		panic(err)
	}
}
