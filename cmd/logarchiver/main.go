package main

import (
	"context"
	"fmt"
	"github.com/TicketsBot/common/observability"
	"github.com/TicketsBot/logarchiver/pkg/config"
	"github.com/TicketsBot/logarchiver/pkg/http"
	"github.com/TicketsBot/logarchiver/pkg/repository"
	"github.com/TicketsBot/logarchiver/pkg/s3client"
	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
	"time"
)

func main() {
	conf := config.Parse[config.Config]()

	if err := sentry.Init(sentry.ClientOptions{
		Dsn:   conf.SentryDsn,
		Debug: !conf.ProductionMode,
	}); err != nil {
		if conf.ProductionMode {
			panic(err)
		} else {
			fmt.Printf("Failed to initialise sentry: %v\n", err)
		}
	}

	var logger *zap.Logger
	var err error
	if conf.ProductionMode {
		logger, err = zap.NewProduction(
			zap.AddCaller(),
			zap.AddStacktrace(zap.ErrorLevel),
			zap.WrapCore(observability.ZapSentryAdapter(observability.EnvironmentProduction)),
		)
	} else {
		logger, err = zap.NewDevelopment(zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	}

	if err != nil {
		panic(err)
	}

	logger.Info("Connecting to database...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	store, err := repository.ConnectPostgres(ctx, conf)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	logger.Info("Connected.")

	logger.Debug("Starting S3 client manager...")
	clientManager := s3client.NewShardedClientManager(conf, store)
	if err := clientManager.Load(ctx); err != nil {
		logger.Fatal("Failed to load S3 clients", zap.Error(err))
	}

	logger.Debug("Starting HTTP server...")

	server := http.NewServer(logger, conf, store, clientManager)
	go server.RemoveQueue.StartReaper()
	server.RegisterRoutes()
	server.Start()
}
