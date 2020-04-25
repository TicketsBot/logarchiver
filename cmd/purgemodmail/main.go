package main

import (
	"flag"
	"github.com/minio/minio-go/v6"
	"strings"
)

var (
	endpoint  = flag.String("endpoint", "nyc3.digitaloceanspaces.com", "the S3 compatible object storage provider endpoint")
	accessKey = flag.String("accesskey", "", "access key ID")
	secretKey = flag.String("secretkey", "", "secret key")
	bucket    = flag.String("bucket", "", "the name of the bucket to manage")
)

func main() {
	flag.Parse()

	client, err := minio.New(*endpoint, *accessKey, *secretKey, false)
	if err != nil {
		panic(err)
	}

	done := make(chan struct{})
	defer close(done)

	for object := range client.ListObjects(*bucket, "", true, done) {
		if strings.Contains(object.Key, "modmail") {
			client.RemoveObject(*bucket, object.Key)
		}
	}
}
