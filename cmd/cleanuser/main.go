package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/TicketsBot/common/encryption"
	"github.com/TicketsBot/logarchiver/pkg/config"
	"github.com/TicketsBot/logarchiver/pkg/model"
	"github.com/TicketsBot/logarchiver/pkg/model/v1"
	"github.com/TicketsBot/logarchiver/pkg/model/v2"
	"github.com/TicketsBot/logarchiver/pkg/s3client"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rxdn/gdl/objects/channel/message"
	"os"
	"strconv"
	"strings"
)

var (
	userId        = flag.Uint64("userid", 0, "user ID to purge")
	guildId       = flag.Uint64("guildid", 0, "guild ID the ticket is from")
	ticketIds     = flag.String("ticket", "", "ticket ID(s) to clean")
	all           = flag.Bool("all", false, "apply to all tickets")
	encryptionKey = flag.String("key", "", "encryption key")
	csv           = flag.String("csv", "", "csv file to read from")
)

func main() {
	flag.Parse()
	cfg := config.Parse()

	// ensure only one is set
	if *ticketIds == "" && !*all && *csv == "" {
		panic("either -ticket, -all or -csv must be set")
	}

	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
	})

	if err != nil {
		panic(err)
	}

	s3Client := s3client.NewS3Client(client, cfg.Bucket)

	var count int
	if *csv != "" {
		tickets := parseCsv(*csv)
		for guildId, ids := range tickets {
			for _, ticketId := range ids {
				msgCount, err := clean(s3Client, guildId, ticketId)
				if err != nil {
					if errors.Is(err, s3client.ErrTicketNotFound) {
						fmt.Printf("ticket %d/%d not found\n", guildId, ticketId)
						continue
					} else {
						panic(err)
					}
				}

				count += msgCount

				fmt.Printf("cleaned %d/%d (%d msgs)\n", guildId, ticketId, msgCount)
			}
		}
	} else if *all {
		keys, err := s3Client.GetAllKeysForGuild(context.Background(), *guildId)
		if err != nil {
			panic(err)
		}

		for _, key := range keys {
			ticketId, err := strconv.Atoi(key[strings.LastIndex(key, "/")+1:])
			if err != nil {
				fmt.Printf("error occurred while parsing id of %s: %v\n", key, err)
				continue
			}

			msgCount, err := clean(s3Client, *guildId, ticketId)
			if err != nil {
				panic(err)
			}

			count += msgCount

			fmt.Printf("cleaned %d\n", ticketId)
		}
	} else {
		split := strings.Split(*ticketIds, ",")
		for _, raw := range split {
			ticketId, err := strconv.Atoi(raw)
			if err != nil {
				panic(err)
			}

			msgCount, err := clean(s3Client, *guildId, ticketId)
			if err != nil {
				panic(err)
			}

			count += msgCount

			fmt.Printf("cleaned ticket %d\n", ticketId)
		}
	}

	fmt.Printf("Cleaned %d messages\n", count)
}

func clean(client *s3client.S3Client, guildId uint64, ticketId int) (int, error) {
	data, err := client.GetTicket(context.Background(), guildId, ticketId)
	if err != nil {
		return 0, err
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
		Id:            0,
		Username:      "Removed for privacy",
		Discriminator: 0,
		Avatar:        "",
		Bot:           false,
	}

	var count int
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

	err = client.StoreTicket(context.Background(), guildId, ticketId, data)
	if err != nil {
		panic(err)
	}

	return count, nil
}

func parseCsv(file string) map[uint64][]int {
	bytes, err := os.ReadFile(file)
	if err != nil {
		panic(err)
	}

	lines := strings.Split(string(bytes), "\n")

	header := strings.Split(lines[0], ",")

	guildIdIdx := -1
	ticketIdIdx := -1

	for i, h := range header {
		if h == "guild_id" {
			guildIdIdx = i
		} else if h == "ticket_id" {
			ticketIdIdx = i
		}
	}

	if guildIdIdx == -1 || ticketIdIdx == -1 {
		panic("Invalid CSV format")
	}

	mapping := make(map[uint64][]int)

	for _, line := range lines[1:] {
		if line == "" {
			continue
		}

		values := strings.Split(line, ",")

		guildId, err := strconv.ParseUint(values[guildIdIdx], 10, 64)
		if err != nil {
			panic(err)
		}

		ticketId, err := strconv.Atoi(values[ticketIdIdx])
		if err != nil {
			panic(err)
		}

		if _, ok := mapping[guildId]; !ok {
			mapping[guildId] = make([]int, 0)
		}

		mapping[guildId] = append(mapping[guildId], ticketId)
	}

	return mapping
}
