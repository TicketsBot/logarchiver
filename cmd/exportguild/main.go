package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/TicketsBot/common/encryption"
	"github.com/TicketsBot/logarchiver/pkg/config"
	"github.com/TicketsBot/logarchiver/pkg/http"
	"github.com/TicketsBot/logarchiver/pkg/model"
	"github.com/TicketsBot/logarchiver/pkg/model/v1"
	v22 "github.com/TicketsBot/logarchiver/pkg/model/v2"
	"github.com/minio/minio-go/v6"
	"github.com/rxdn/gdl/objects/channel/message"
	"os"
	"strconv"
	"strings"
)

var (
	guildId       = flag.Uint64("guildid", 0, "guild id to export")
	key           = flag.String("key", "", "aes key")
	ticketId      = flag.Int("ticketid", 0, "set to export a single ticket")
	convert       = flag.Bool("convert", false, "convert to v2 if necessary")
	userWhitelist = flag.Uint64("userwhitelist", 0, "only export tickets from this user")
)

func main() {
	flag.Parse()
	conf := config.Parse()

	// create minio client
	client, err := minio.New(conf.Endpoint, conf.AccessKey, conf.SecretKey, false)
	if err != nil {
		panic(err)
	}

	s := http.NewServer(conf, client)

	// likely to be file exists
	_ = os.Mkdir(fmt.Sprintf("export/%d", *guildId), 0)

	if ticketId != nil && *ticketId > 0 {
		export(*ticketId, s)
	} else {
		doneCh := make(chan struct{})
		defer close(doneCh)

		objCh := client.ListObjectsV2(conf.Bucket, fmt.Sprintf("%d/", *guildId), true, doneCh)

		for obj := range objCh {
			id := obj.Key
			id = strings.Replace(id, fmt.Sprintf("%d/", *guildId), "", -1)
			id = strings.Replace(id, "free-", "", -1)
			parsed, err := strconv.Atoi(id)
			must(err)

			export(parsed, s)
		}
	}
}

func export(id int, s *http.Server) {
	data, err := s.GetTicket(s.Config.Bucket, *guildId, id)
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
