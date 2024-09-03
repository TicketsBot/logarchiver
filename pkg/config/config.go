package config

import (
	"github.com/caarlos0/env"
	"github.com/google/uuid"
	"reflect"
)

type Config struct {
	Address string `env:"ARCHIVER_ADDR"`

	AccessKey       string    `env:"S3_ACCESS"`
	SecretKey       string    `env:"S3_SECRET"`
	DefaultBucketId uuid.UUID `env:"DEFAULT_BUCKET_ID"`

	SentryDsn      string `env:"SENTRY_DSN"`
	ProductionMode bool   `env:"PRODUCTION_MODE" envDefault:"false"`
	AdminAuthToken string `env:"ADMIN_AUTH_TOKEN"`

	DatabaseUri string `env:"DATABASE_URI"`
}

func Parse() (conf Config) {
	parsers := env.CustomParsers{
		reflect.TypeOf(uuid.UUID{}): func(value string) (interface{}, error) {
			return uuid.Parse(value)
		},
	}

	if err := env.ParseWithFuncs(&conf, parsers); err != nil {
		panic(err)
	}

	return
}
