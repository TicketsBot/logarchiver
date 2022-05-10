package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/TicketsBot/common/encryption"
	"github.com/TicketsBot/logarchiver/config"
	"github.com/TicketsBot/logarchiver/http"
	"github.com/minio/minio-go/v6"
	"os"
	"strconv"
	"strings"
)

var (
	guildId  = flag.Uint64("guildid", 0, "guild id to export")
	key      = flag.String("key", "", "aes key")
	ticketId = flag.Int("ticketid", 0, "set to export a single ticket")
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
