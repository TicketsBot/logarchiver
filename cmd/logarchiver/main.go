package main

import (
	"github.com/TicketsBot/logarchiver/config"
	"github.com/TicketsBot/logarchiver/http"
	"github.com/minio/minio-go/v6"
)

func main() {
	config.LoadConfig()

	// create minio client
	client, err := minio.New(config.Conf.S3.Endpoint, config.Conf.S3.AccessKey, config.Conf.S3.SecretKey, false)
	if err != nil {
		panic(err)
	}
	
	server := http.NewServer(client)
	server.RegisterRoutes()
	server.Start()
}
