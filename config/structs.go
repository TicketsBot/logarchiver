package config

type (
	Config struct {
		Address string
		S3      S3
	}

	S3 struct {
		Endpoint  string
		AccessKey string
		SecretKey string
		Bucket    string
	}
)
