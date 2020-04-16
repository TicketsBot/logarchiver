package logarchiver

import "github.com/TicketsBot/logarchiver/http"

type LogArchiver struct {
	HttpServer *http.HttpServer
}

func NewLogArchiver() *LogArchiver {
	return &LogArchiver{
		HttpServer: http.NewHttpServer(),
	}
}
