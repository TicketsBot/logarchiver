package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/TicketsBot/common/encryption"
	"github.com/TicketsBot/logarchiver/discord"
	"github.com/TicketsBot/logarchiver/http"
	"github.com/minio/minio-go/v6"
	"strconv"
	"strings"
)

var (
	endpoint  = flag.String("endpoint", "nyc3.digitaloceanspaces.com", "the S3 compatible object storage provider endpoint")
	accessKey = flag.String("accesskey", "", "access key ID")
	secretKey = flag.String("secretkey", "", "secret key")
	bucket    = flag.String("bucket", "", "the name of the bucket to manage")

	userId = flag.Uint64("userid", 0, "user ID to purge")
	guildId = flag.Uint64("guildid", 0, "guild ID the ticket is from")
	ticketId = flag.Int("ticket", 0, "ticket ID to clean")
	all = flag.Bool("all", false, "apply to all tickets")
	encryptionKey = flag.String("key", "eo#6dDqK6&!G1OA$EqBYKr4l2^PrT^Bp", "encryption key")
)

func main() {
	flag.Parse()

	if *ticketId == 0 && !*all || *ticketId > 1 && *all {
		panic("ticket or all must be set and are mutually exclusive")
	}

	client, err := minio.New(*endpoint, *accessKey, *secretKey, false)
	if err != nil {
		panic(err)
	}

	server := http.NewServer(client)
	if !*all {
		clean(server, *ticketId)
	} else {
		done := make(chan struct{})
		defer close(done)

		prefix := fmt.Sprintf("%d/", *guildId)
		for obj := range client.ListObjectsV2WithMetadata(*bucket, prefix, true, done) {
			suffix := strings.TrimPrefix(obj.Key, prefix)
			suffix = strings.TrimPrefix(suffix, "free-")
			ticketId, err := strconv.Atoi(suffix)
			if err != nil {
				fmt.Printf("error occurred while parsing id of %s: %v\n", obj.Key, err)
				continue
			}

			clean(server, ticketId)
			fmt.Printf("cleaned %d\n", ticketId)
		}
	}
}

func clean(server *http.Server, ticketId int) {
	data, isPremium, err := server.GetTicket(*bucket, *guildId, ticketId)
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

	var messages []discord.Message
	if err := json.Unmarshal(data, &messages); err != nil {
		panic(err)
	}

	for i, message := range messages {
		if message.Author.Id == *userId {
			message.Author.Username = strconv.FormatUint(message.Author.Id, 10)
			message.Author.Avatar = ""
			message.Content = "Deleted upon request of user"
			message.Attachments = nil

			messages[i] = message
		}
	}

	data, err = json.Marshal(messages)
	if err != nil {
		panic(err)
	}

	data, err = encryption.Encrypt([]byte(*encryptionKey), data)
	if err != nil {
		panic(err)
	}

	data = encryption.Compress(data)

	err = server.UploadTicket(*bucket, isPremium, *guildId, ticketId, data)
	if err != nil {
		panic(err)
	}
}
