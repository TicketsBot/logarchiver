package http

import (
	"github.com/TicketsBot/logarchiver/config"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v6"
	"log"
	"os"
)

type Server struct {
	Logger *log.Logger
	Config config.Config
	router *gin.Engine
	client *minio.Client
}

func NewServer(config config.Config, client *minio.Client) *Server {
	return &Server{
		Logger: log.New(os.Stdout, "[server] ", log.LstdFlags),
		Config: config,
		router: gin.Default(),
		client: client,
	}
}

func (s *Server) RegisterRoutes() {
	s.router.LoadHTMLGlob("./public/templates/*")

	s.router.POST("/encode", encodeHandler)

	s.router.GET("/", s.ticketGetHandler)
	s.router.POST("/", s.ticketUploadHandler)

	s.router.DELETE("/guild/:id", s.purgeGuildHandler)
}

func (s *Server) Start() {
	if err := s.router.Run(s.Config.Address); err != nil {
		panic(err)
	}
}
