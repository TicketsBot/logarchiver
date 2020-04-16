package main

import "github.com/TicketsBot/logarchiver"

func main() {
	archiver := logarchiver.NewLogArchiver()
	archiver.HttpServer.RegisterRoutes()
	archiver.HttpServer.Start()
}
