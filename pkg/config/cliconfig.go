package config

type CliConfig struct {
	AccessKey string `env:"S3_ACCESS,required"`
	SecretKey string `env:"S3_SECRET,required"`
	Endpoint  string `env:"S3_ENDPOINT,required"`
	Bucket    string `env:"S3_BUCKET,required"`
}
