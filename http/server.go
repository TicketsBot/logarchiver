package http

import (
	"github.com/TicketsBot/logarchiver/config"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v6"
)

type Server struct {
	router *gin.Engine
	client *minio.Client
}

func NewServer(client *minio.Client) *Server {
	return &Server{
		router: gin.Default(),
		client: client,
	}
}


func (s *Server) RegisterRoutes() {
	s.router.LoadHTMLGlob("./public/templates/*")

	s.router.POST("/encode", encodeHandler)

	s.router.GET("/", s.ticketGetHandler)
	s.router.POST("/", s.ticketUploadHandler)

	s.router.GET("/modmail", s.modmailGetHandler)
	s.router.POST("/modmail", s.modmailUploadHandler)

	s.router.GET("/modmail/all", s.modmailListHandler)
}

func (s *Server) Start() {
	if err := s.router.Run(config.Conf.Address); err != nil {
		panic(err)
	}
}
