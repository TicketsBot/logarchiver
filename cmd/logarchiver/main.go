package main

import (
	"bytes"
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

	client.PutObject(config.Conf.S3.Bucket, "yep/free-file", bytes.NewReader([]byte("yep")), 3, minio.PutObjectOptions{
	})
	
	server := http.NewServer(client)
	server.RegisterRoutes()
	server.Start()
}
