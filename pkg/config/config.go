package config

import "github.com/caarlos0/env"

type Config struct {
	Address string `env:"ARCHIVER_ADDR"`

	Endpoint  string `env:"S3_ENDPOINT"`
	Bucket    string `env:"S3_BUCKET"`
	AccessKey string `env:"S3_ACCESS"`
	SecretKey string `env:"S3_SECRET"`

	SentryDsn      string `env:"SENTRY_DSN"`
	ProductionMode bool   `env:"PRODUCTION_MODE" envDefault:"false"`
}

func Parse() (conf Config) {
	if err := env.Parse(&conf); err != nil {
		panic(err)
	}

	return
}
