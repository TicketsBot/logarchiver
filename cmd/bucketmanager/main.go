package main

import (
	"bytes"
	"encoding/xml"
	"flag"
	"github.com/minio/minio-go/v6"
)

var (
	endpoint  = flag.String("endpoint", "nyc3.digitaloceanspaces.com", "the S3 compatible object storage provider endpoint")
	accessKey = flag.String("accesskey", "", "access key ID")
	secretKey = flag.String("secretkey", "", "secret key")
	bucket    = flag.String("bucket", "", "the name of the bucket to manage")

	rule   = flag.String("rule", "", "A unique string identifying the rule. It may contain up to 255 characters including spaces")
	prefix = flag.String("prefix", "", "Prefix of files to apply the rule to")
	status = flag.Bool("enabled", false, "Whether or not the rule will be enabled")
	days   = flag.Int("expiration.days", 90, "An integer specifying the number of days after an object's creation until the rule takes effect")
)

func main() {
	flag.Parse()

	client, err := minio.New(*endpoint, *accessKey, *secretKey, false)
	if err != nil {
		panic(err)
	}

	var buff bytes.Buffer
	encoder := xml.NewEncoder(&buff)
	encoder.Indent("  ", "    ")
	if err := encoder.Encode(NewLifecycleConfiguration(*rule, *prefix, *status, *days)); err != nil {
		panic(err)
	}

	if err := client.SetBucketLifecycle(*bucket, string(buff.Bytes())); err != nil {
		panic(err)
	}
}
