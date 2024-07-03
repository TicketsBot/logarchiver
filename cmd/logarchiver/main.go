package main

import (
	"fmt"
	"github.com/TicketsBot/common/observability"
	"github.com/TicketsBot/logarchiver/pkg/config"
	"github.com/TicketsBot/logarchiver/pkg/http"
	"github.com/getsentry/sentry-go"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

func main() {
	conf := config.Parse()

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

	logger.Debug("Starting minio client...")

	// create minio client
	client, err := minio.New(conf.Endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(conf.AccessKey, conf.SecretKey, ""),
	})
	if err != nil {
		logger.Fatal("Failed to create minio client", zap.Error(err), zap.String("endpoint", conf.Endpoint))
		panic(err) // logger.Fatal should exit already
	}

	logger.Debug("Starting HTTP server...")

	server := http.NewServer(logger, conf, client)
	go server.RemoveQueue.StartReaper()
	server.RegisterRoutes()
	server.Start()
}
