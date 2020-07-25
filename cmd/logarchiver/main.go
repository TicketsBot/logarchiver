package main

import (
	"github.com/TicketsBot/logarchiver/http"
	"github.com/minio/minio-go/v6"
	"os"
)

func main() {
	// create minio client
	client, err := minio.New(os.Getenv("S3_ENDPOINT"), os.Getenv("S3_ACCESS"), os.Getenv("S3_SECRET"), false)
	if err != nil {
		panic(err)
	}
	
	server := http.NewServer(client)
	server.RegisterRoutes()
	server.Start()
}
