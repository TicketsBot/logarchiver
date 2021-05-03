package main

import (
	"github.com/TicketsBot/logarchiver/config"
	"github.com/TicketsBot/logarchiver/http"
	"github.com/minio/minio-go/v6"
)

func main() {
	conf := config.Parse()

	// create minio client
	client, err := minio.New(conf.Endpoint, conf.AccessKey, conf.SecretKey, false)
	if err != nil {
		panic(err)
	}

	server := http.NewServer(conf, client)
	server.RegisterRoutes()
	server.Start()
}
