package http

import (
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v6"
	"os"
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
}

func (s *Server) Start() {
	if err := s.router.Run(os.Getenv("ARCHIVER_ADDR")); err != nil {
		panic(err)
	}
}
