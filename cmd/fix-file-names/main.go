package main

import (
	"context"
	"flag"
	"github.com/TicketsBot/logarchiver/pkg/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
	"strings"
	"sync"
	"sync/atomic"
)

const workers = 30

func main() {
	flag.Parse()
	conf := config.Parse()

	// create minio client
	client, err := minio.New(conf.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(conf.AccessKey, conf.SecretKey, ""),
		Secure: true,
	})
	if err != nil {
		panic(err)
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	ch := client.ListObjects(context.Background(), conf.Bucket, minio.ListObjectsOptions{
		Prefix:    "8",
		Recursive: true,
	})

	var wg sync.WaitGroup
	processed := atomic.Int32{}
	freeCount := atomic.Int32{}
	for i := 0; i < workers; i++ {
		wg.Add(1)

		go func() {
			for obj := range ch {
				if strings.Contains(obj.Key, "free-") {
					newName := strings.Replace(obj.Key, "free-", "", 1)

					src := minio.CopySrcOptions{
						Bucket: conf.Bucket,
						Object: obj.Key,
					}

					dst := minio.CopyDestOptions{
						Bucket: conf.Bucket,
						Object: newName,
					}

					if _, err := client.CopyObject(context.Background(), dst, src); err != nil {
						logger.Fatal("Failed to copy object", zap.Error(err))
						continue
					}

					if err := client.RemoveObject(context.Background(), conf.Bucket, obj.Key, minio.RemoveObjectOptions{}); err != nil {
						logger.Fatal("Failed to remove object", zap.Error(err))
						continue
					}

					updated := freeCount.Add(1)
					logger.Info("Processed free ticket", zap.String("key", obj.Key), zap.Int32("processed", updated))
				}

				updated := processed.Add(1)
				if updated%100_000 == 0 {
					logger.Info("Processed objects", zap.Int32("processed", updated))
				}
			}

			wg.Done()
		}()
	}

	wg.Wait()
}
