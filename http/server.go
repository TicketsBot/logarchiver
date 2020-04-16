package http

import (
	"github.com/TicketsBot/logarchiver/http/routes"
	"github.com/gin-gonic/gin"
)

type HttpServer struct {
	router *gin.Engine
}

func NewHttpServer() *HttpServer {
	return &HttpServer{
		router: gin.Default(),
	}
}

func (s *HttpServer) RegisterRoutes() {
	s.router.LoadHTMLGlob("./public/templates/*")

	s.router.POST("/encode", routes.EncodeHandler)
}

func (s *HttpServer) Start() {
	if err := s.router.Run(":3000"); err != nil {
		panic(err)
	}
}
