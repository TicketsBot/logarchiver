package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/TicketsBot/common/encryption"
	"github.com/TicketsBot/logarchiver/config"
	"github.com/TicketsBot/logarchiver/http"
	"github.com/TicketsBot/logarchiver/model"
	"github.com/TicketsBot/logarchiver/model/v1"
	v2 "github.com/TicketsBot/logarchiver/model/v2"
	"github.com/minio/minio-go/v6"
	"github.com/rxdn/gdl/objects/channel/message"
	"os"
	"strconv"
	"strings"
)

var (
	endpoint, accessKey, secretKey, bucket string

	userId        = flag.Uint64("userid", 0, "user ID to purge")
	guildId       = flag.Uint64("guildid", 0, "guild ID the ticket is from")
	ticketId      = flag.Int("ticket", 0, "ticket ID to clean")
	all           = flag.Bool("all", false, "apply to all tickets")
	encryptionKey = flag.String("key", "", "encryption key") // to any keen eyes looking at commit history: the key has been ommitted and all transcripts have been re-encrypted
)

func loadEnvvars() {
	endpoint = strip(os.Getenv("S3_ENDPOINT"))
	accessKey = strip(os.Getenv("S3_ACCESS"))
	secretKey = strip(os.Getenv("S3_SECRET"))
	bucket = strip(os.Getenv("S3_BUCKET"))
}

func strip(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\r", ""), "\n", "")
}

func main() {
	flag.Parse()
	loadEnvvars()

	if *ticketId == 0 && !*all || *ticketId > 1 && *all {
		panic("ticket or all must be set and are mutually exclusive")
	}

	conf := config.Config{
		Endpoint:  endpoint,
		Bucket:    bucket,
		AccessKey: accessKey,
		SecretKey: secretKey,
	}

	client, err := minio.New(endpoint, accessKey, secretKey, false)
	if err != nil {
		panic(err)
	}

	server := http.NewServer(conf, client)

	var count int
	if !*all {
		count = clean(server, *ticketId)
	} else {
		done := make(chan struct{})
		defer close(done)

		prefix := fmt.Sprintf("%d/", *guildId)
		for obj := range client.ListObjectsV2WithMetadata(bucket, prefix, true, done) {
			suffix := strings.TrimPrefix(obj.Key, prefix)
			suffix = strings.TrimPrefix(suffix, "free-")
			ticketId, err := strconv.Atoi(suffix)
			if err != nil {
				fmt.Printf("error occurred while parsing id of %s: %v\n", obj.Key, err)
				continue
			}

			count += clean(server, ticketId)
			fmt.Printf("cleaned %d\n", ticketId)
		}
	}

	fmt.Printf("Cleaned %d messages\n", count)
}

func clean(server *http.Server, ticketId int) (count int) {
	data, err := server.GetTicket(bucket, *guildId, ticketId)
	if err != nil {
		panic(err)
	}

	data, err = encryption.Decompress(data)
	if err != nil {
		panic(err)
	}

	data, err = encryption.Decrypt([]byte(*encryptionKey), data)
	if err != nil {
		panic(err)
	}

	var transcript v2.Transcript

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

	transcript.Entities.Users[*userId] = v2.User{
		Id:            *userId,
		Username:      strconv.FormatUint(*userId, 10),
		Discriminator: 0,
		Avatar:        "",
		Bot:           false,
	}

	for i, message := range transcript.Messages {
		if message.AuthorId == *userId {
			count++

			message.AuthorId = 0
			message.Content = "Deleted upon request of user"
			message.Embeds = nil
			message.Attachments = nil

			transcript.Messages[i] = message
		}
	}

	data, err = json.Marshal(transcript)
	if err != nil {
		panic(err)
	}

	data, err = encryption.Encrypt([]byte(*encryptionKey), data)
	if err != nil {
		panic(err)
	}

	data = encryption.Compress(data)

	err = server.UploadTicket(bucket, *guildId, ticketId, data)
	if err != nil {
		panic(err)
	}

	return
}
