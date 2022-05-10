package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/TicketsBot/common/encryption"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/logarchiver/config"
	"github.com/TicketsBot/logarchiver/http"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/minio/minio-go/v6"
	"os"
)

var (
	userId         = flag.Uint64("userid", 0, "user id to export")
	key            = flag.String("key", "", "aes key")
	fullTranscript = flag.Bool("fulltranscript", false, "export full transcript") // TODO: Implement
	dbUri          = flag.String("dburi", "", "database uri")
	cacheUri       = flag.String("cacheuri", "", "cache uri")
)

func main() {
	flag.Parse()
	conf := config.Parse()

	// likely to be file exists
	_ = os.Mkdir(fmt.Sprintf("export_user/%d", *userId), 0)

	var db *database.Database
	{
		pool, err := pgxpool.Connect(context.Background(), *dbUri)
		must(err)

		db = database.NewDatabase(pool)
	}

	/*
		var c cache.PgCache
		{
			pool, err := pgxpool.Connect(context.Background(), *cacheUri)
			must(err)

			c = cache.NewPgCache(pool, cache.CacheOptions{
				Users:   true,
				Members: true,
			})
		}
	*/

	// Get + write user data
	/*
		{
			userData := getUserData(db, *userId)
			encoded, err := json.MarshalIndent(userData, "", "  ")
			must(err)

			writeFile(fmt.Sprintf("export_user/%d/database.json", *userId), encoded)
		}
	*/

	/*
		{
			cacheData := getCacheData(&c, *userId)

			encoded, err := json.MarshalIndent(cacheData, "", "  ")
			must(err)

			writeFile(fmt.Sprintf("export_user/%d/cache.json", *userId), encoded)
		}
	*/

	transcriptIds := make(map[uint64][]int)

	{
		query := `SELECT participant.guild_id, participant.ticket_id FROM participant INNER JOIN tickets ON participant.guild_id = tickets.guild_id AND tickets.id = participant.ticket_id WHERE participant.user_id = $1 AND tickets.has_transcript='t' and tickets.open='f';`
		rows, err := db.Participants.Query(context.Background(), query, *userId)
		must(err)

		for rows.Next() {
			var guildId uint64
			var ticketId int

			must(rows.Scan(&guildId, &ticketId))

			if transcriptIds[guildId] == nil {
				transcriptIds[guildId] = make([]int, 0)
			}

			transcriptIds[guildId] = append(transcriptIds[guildId], ticketId)
		}
	}

	{
		query := `SELECT guild_id, id FROM tickets WHERE user_id = $1 AND has_transcript='t' AND open='f';`
		rows, err := db.Tickets.Query(context.Background(), query, *userId)
		must(err)

		for rows.Next() {
			var guildId uint64
			var ticketId int

			must(rows.Scan(&guildId, &ticketId))

			if transcriptIds[guildId] == nil {
				transcriptIds[guildId] = make([]int, 0)
			}

			transcriptIds[guildId] = append(transcriptIds[guildId], ticketId)
		}
	}

	getTranscripts(conf, transcriptIds)
}

func getTranscripts(conf config.Config, tickets map[uint64][]int) {
	// create minio client
	client, err := minio.New(conf.Endpoint, conf.AccessKey, conf.SecretKey, false)
	if err != nil {
		panic(err)
	}

	s := http.NewServer(conf, client)

	doneCh := make(chan struct{})
	defer close(doneCh)

	_ = os.Mkdir(fmt.Sprintf("export_user/%d/transcripts", *userId), 0)

	for guildId, ticketIds := range tickets {
		for _, ticketId := range ticketIds {
			data, err := s.GetTicket(conf.Bucket, guildId, ticketId)
			must(err)

			data, err = encryption.Decompress(data)
			must(err)

			data, err = encryption.Decrypt([]byte(*key), data)
			must(err)

			var encoded bytes.Buffer
			must(json.Indent(&encoded, data, "", "  "))

			f, err := os.Create(fmt.Sprintf("export_user/%d/transcripts/%d-%d.json", *userId, guildId, ticketId))
			must(err)

			_, err = encoded.WriteTo(f)
			must(err)

			fmt.Printf("exported %d/%d\n", guildId, ticketId)
		}
	}
}

func writeFile(fileName string, data []byte) {
	f, err := os.Create(fileName)
	must(err)

	_, err = f.Write(data)
	must(err)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
