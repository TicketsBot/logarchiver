package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/TicketsBot/common/encryption"
	"github.com/TicketsBot/logarchiver/pkg/config"
	"github.com/TicketsBot/logarchiver/pkg/model"
	"github.com/TicketsBot/logarchiver/pkg/model/v1"
	v22 "github.com/TicketsBot/logarchiver/pkg/model/v2"
	"github.com/TicketsBot/logarchiver/pkg/s3client"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rxdn/gdl/objects/channel/message"
	"golang.org/x/sync/errgroup"
	"os"
	"strconv"
	"strings"
)

const workers = 15

var (
	guildId       = flag.Uint64("guildid", 0, "guild id to export")
	key           = flag.String("key", "", "aes key")
	ticketId      = flag.Int("ticketid", 0, "set to export a single ticket")
	convert       = flag.Bool("convert", false, "convert to v2 if necessary")
	userWhitelist = flag.Uint64("userwhitelist", 0, "only export tickets from this user")
	after         = flag.Int("after", 0, "export ticket IDs above this value (inclusive)")
)

func main() {
	flag.Parse()
	conf := config.Parse()

	// create minio client
	m, err := minio.New(conf.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(conf.AccessKey, conf.SecretKey, ""),
		Secure: true,
	})
	if err != nil {
		panic(err)
	}

	client := s3client.NewS3Client(m, conf.Bucket)

	// likely to be file exists
	_ = os.Mkdir(fmt.Sprintf("export/%d", *guildId), 0)

	if ticketId != nil && *ticketId > 0 {
		export(*ticketId, client)
	} else {
		keys, err := client.GetAllKeysForGuild(context.Background(), *guildId)
		if err != nil {
			panic(err)
		}

		keyCh := make(chan string)
		go func() {
			for _, key := range keys {
				keyCh <- key
			}

			close(keyCh)
		}()

		group, _ := errgroup.WithContext(context.Background())
		for i := 0; i < workers; i++ {
			group.Go(func() error {
				for key := range keyCh {
					id := key[strings.LastIndex(key, "/")+1:]
					parsed, err := strconv.Atoi(id)
					must(err)

					if after != nil && *after > 0 && parsed < *after {
						continue
					}

					export(parsed, client)
				}

				return nil
			})
		}

		if err := group.Wait(); err != nil {
			panic(err)
		}
	}
}

func export(id int, client *s3client.S3Client) {
	data, err := client.GetTicket(context.Background(), *guildId, id)
	must(err)

	data, err = encryption.Decompress(data)
	must(err)

	data, err = encryption.Decrypt([]byte(*key), data)
	must(err)

	if *convert || (userWhitelist != nil && *userWhitelist > 0) {
		var transcript v22.Transcript

		version := model.GetVersion(data)
		switch version {
		case model.V1:
			var messages []message.Message
			if err := json.Unmarshal(data, &messages); err != nil {
				panic(err)
			}

			transcript = v1.ConvertToV2(messages)
		case model.V2:
			if err := json.Unmarshal(data, &transcript); err != nil {
				panic(err)
			}
		default:
			panic(fmt.Sprintf("Unknown version %d", version))
		}

		data, err = json.Marshal(transcript)
		must(err)
	}

	if userWhitelist != nil && *userWhitelist > 0 {
		var transcript v22.Transcript
		if err := json.Unmarshal(data, &transcript); err != nil {
			panic(err)
		}

		transcript.Entities.Channels = nil
		transcript.Entities.Roles = nil

		user, ok := transcript.Entities.Users[*userWhitelist]
		if !ok {
			transcript.Entities.Users = nil
		} else {
			transcript.Entities.Users = map[uint64]v22.User{
				user.Id: user,
			}
		}

		var messages []v22.Message
		for _, message := range transcript.Messages {
			if message.AuthorId == *userWhitelist {
				messages = append(messages, message)
			}
		}

		transcript.Messages = messages

		data, err = json.Marshal(transcript)
		must(err)
	}

	var encoded bytes.Buffer
	must(json.Indent(&encoded, data, "", "  "))

	f, err := os.Create(fmt.Sprintf("export/%d/%d.json", *guildId, id))
	must(err)

	_, err = encoded.WriteTo(f)
	must(err)

	fmt.Printf("exported %d\n", id)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
