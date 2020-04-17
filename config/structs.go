package config

type (
	Config struct {
		S3 S3
	}

	S3 struct {
		Endpoint  string
		AccessKey string
		SecretKey string
		Bucket    string
	}
)
