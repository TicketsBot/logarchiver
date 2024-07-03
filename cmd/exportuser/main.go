package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/TicketsBot/common/encryption"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/logarchiver/pkg/config"
	"github.com/TicketsBot/logarchiver/pkg/model"
	v1 "github.com/TicketsBot/logarchiver/pkg/model/v1"
	v2 "github.com/TicketsBot/logarchiver/pkg/model/v2"
	"github.com/TicketsBot/logarchiver/pkg/s3client"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rxdn/gdl/cache"
	"github.com/rxdn/gdl/objects/channel/message"
	"go.uber.org/zap"
	"os"
	"strconv"
	"time"
)

var (
	userId   = flag.Uint64("userid", 0, "user id to export")
	key      = flag.String("key", "", "aes key")
	dbUri    = flag.String("dburi", "", "database uri")
	cacheUri = flag.String("cacheuri", "", "cache uri")
)

func main() {
	flag.Parse()
	conf := config.Parse()

	// likely to be file exists
	_ = os.Mkdir(fmt.Sprintf("export_user/%d", *userId), 0)

	fmt.Println("Connecting to database...")

	var db *database.Database
	{
		pool, err := pgxpool.Connect(context.Background(), *dbUri)
		must(err)

		db = database.NewDatabase(pool)
	}

	fmt.Println("Connecting to cache...")

	var c cache.PgCache
	{
		pool, err := pgxpool.Connect(context.Background(), *cacheUri)
		must(err)

		c = cache.NewPgCache(pool, cache.CacheOptions{
			Users:   true,
			Members: true,
		})
	}

	fmt.Println("Connected to cache")

	// Get + write user data
	{
		userData := getUserData(db, *userId)
		encoded, err := json.MarshalIndent(userData, "", "  ")
		must(err)

		writeFile(fmt.Sprintf("export_user/%d/database.json", *userId), encoded)
	}

	{
		cacheData := getCacheData(&c, *userId)

		encoded, err := json.MarshalIndent(cacheData, "", "  ")
		must(err)

		writeFile(fmt.Sprintf("export_user/%d/cache.json", *userId), encoded)
	}

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
	m, err := minio.New(conf.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(conf.AccessKey, conf.SecretKey, ""),
		Secure: true,
	})
	if err != nil {
		panic(err)
	}

	client := s3client.NewS3Client(m, conf.Bucket)

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	_ = os.Mkdir(fmt.Sprintf("export_user/%d/transcripts", *userId), 0)

	for guildId, ticketIds := range tickets {
		for _, ticketId := range ticketIds {
			data, err := client.GetTicket(context.Background(), guildId, ticketId)
			if err != nil {
				if errors.Is(err, s3client.ErrTicketNotFound) {
					logger.Info("ticket not found", zap.Uint64("guildId", guildId), zap.Int("ticketId", ticketId))
					continue
				} else {
					panic(err)
				}
			}

			data, err = encryption.Decompress(data)
			if err != nil {
				logger.Error("failed to decompress", zap.Error(err), zap.Uint64("guildId", guildId), zap.Int("ticketId", ticketId))
				continue
			}

			data, err = encryption.Decrypt([]byte(*key), data)
			if err != nil {
				logger.Error("failed to decrypt", zap.Error(err), zap.Uint64("guildId", guildId), zap.Int("ticketId", ticketId))
				continue
			}

			// Convert to v2 if needed
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

			transcript.Entities.Channels = nil
			transcript.Entities.Roles = nil

			user, ok := transcript.Entities.Users[*userId]
			if !ok {
				transcript.Entities.Users = nil
			} else {
				transcript.Entities.Users = map[uint64]v2.User{
					user.Id: user,
				}
			}

			var messages []v2.Message
			for _, message := range transcript.Messages {
				if message.AuthorId == *userId {
					messages = append(messages, message)
				}
			}

			transcript.Messages = messages

			encoded, err := json.MarshalIndent(transcript, "", "  ")
			must(err)

			fileName := fmt.Sprintf("export_user/%d/transcripts/%d-%d.json", *userId, guildId, ticketId)
			must(os.WriteFile(fileName, encoded, 0644))

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

// user data

// The worst code you have ever seen
func getUserData(db *database.Database, userId uint64) map[string]interface{} {
	data := make(map[string]interface{})

	data["blacklisted_guilds"] = getBlacklistedGuilds(db, userId)
	data["close_requests"] = getCloseRequests(db, userId)
	data["response_times"] = getResponseTimes(db, userId)
	data["participated_tickets"] = getParticipatedTickets(db, userId)
	data["permissions"] = getPermissions(db, userId)
	data["team_permissions"] = getTeamPermissions(db, userId)
	data["claimed_tickets"] = getClaimedTickets(db, userId)
	data["member_of_tickets"] = getTicketsMember(db, userId)
	data["tickets"] = getTickets(db, userId)
	data["premium_activated_for"] = getPremiumActivatedFor(db, userId)

	guilds, err := db.UserGuilds.Get(userId)
	must(err)
	data["guilds"] = guilds

	voteTime, err := db.Votes.Get(userId)
	must(err)
	if voteTime.IsZero() {
		data["last_vote_time"] = nil
	} else {
		data["last_vote_time"] = voteTime
	}

	whitelabel, err := db.Whitelabel.GetByUserId(userId)
	must(err)
	if whitelabel.UserId == 0 {
		data["whitelabel"] = nil
	} else {
		data["whitelabel"] = whitelabel
	}

	whitelabelExpiry, err := db.WhitelabelUsers.GetExpiry(userId)
	must(err)
	if whitelabelExpiry.IsZero() {
		data["whitelabel_expiry"] = nil
	} else {
		data["whitelabel_expiry"] = whitelabelExpiry
	}

	return data
}

func getBlacklistedGuilds(db *database.Database, userId uint64) (guilds []uint64) {
	rows, err := db.Blacklist.Query(context.Background(), "SELECT guild_id FROM blacklist WHERE user_id = $1;", userId)
	must(err)

	for rows.Next() {
		var guildId uint64
		must(rows.Scan(&guildId))

		guilds = append(guilds, guildId)
	}

	return
}

func getCloseRequests(db *database.Database, userId uint64) (requests []database.CloseRequest) {
	query := `
SELECT "guild_id", "ticket_id", "user_id", "close_at", "close_reason"
FROM close_request
WHERE "user_id" = $1;`

	rows, err := db.Blacklist.Query(context.Background(), query, userId)
	must(err)

	for rows.Next() {
		var request database.CloseRequest
		must(rows.Scan(&request.GuildId, &request.TicketId, &request.UserId, &request.CloseAt, &request.Reason))
		requests = append(requests, request)
	}

	return
}

func getResponseTimes(db *database.Database, userId uint64) (times []interface{}) {
	query := `
SELECT "guild_id", "ticket_id", "user_id", "response_time"
FROM first_response_time
WHERE "user_id" = $1;`

	rows, err := db.Blacklist.Query(context.Background(), query, userId)
	must(err)

	for rows.Next() {
		var guildId, userId uint64
		var ticketId int
		var responseTime time.Duration

		must(rows.Scan(&guildId, &ticketId, &userId, &responseTime))
		times = append(times, map[string]interface{}{
			"guild_id":      guildId,
			"ticket_id":     ticketId,
			"user_id":       userId,
			"response_time": responseTime,
		})
	}

	return
}

func getParticipatedTickets(db *database.Database, userId uint64) (tickets []string) {
	query := `
SELECT "guild_id", "ticket_id"
FROM participant
WHERE "user_id" = $1;`

	rows, err := db.Blacklist.Query(context.Background(), query, userId)
	must(err)

	for rows.Next() {
		var guildId uint64
		var ticketId int

		must(rows.Scan(&guildId, &ticketId))
		tickets = append(tickets, fmt.Sprintf("%d/%d", guildId, ticketId))
	}

	return
}

func getPermissions(db *database.Database, userId uint64) map[uint64]string {
	query := `
SELECT "guild_id", "support", "admin"
FROM permissions
WHERE "user_id" = $1;`

	rows, err := db.Blacklist.Query(context.Background(), query, userId)
	must(err)

	data := make(map[uint64]string)
	for rows.Next() {
		var guildId uint64
		var isSupport, isAdmin bool

		must(rows.Scan(&guildId, &isSupport, &isAdmin))

		if isAdmin {
			data[guildId] = "admin"
		} else if isSupport {
			data[guildId] = "support"
		} else {
			data[guildId] = "none"
		}
	}

	return data
}

func getTeamPermissions(db *database.Database, userId uint64) (teams []int) {
	query := `
SELECT "team_id"
FROM support_team_members
WHERE "user_id" = $1;`

	rows, err := db.Blacklist.Query(context.Background(), query, userId)
	must(err)

	for rows.Next() {
		var teamId int
		must(rows.Scan(&teamId))
		teams = append(teams, teamId)
	}

	return
}

func getClaimedTickets(db *database.Database, userId uint64) (tickets []string) {
	query := `
SELECT "guild_id", "ticket_id"
FROM ticket_claims
WHERE "user_id" = $1;`

	rows, err := db.Blacklist.Query(context.Background(), query, userId)
	must(err)

	for rows.Next() {
		var guildId uint64
		var ticketId int

		must(rows.Scan(&guildId, &ticketId))
		tickets = append(tickets, fmt.Sprintf("%d/%d", guildId, ticketId))
	}

	return
}

func getTicketsMember(db *database.Database, userId uint64) (tickets []string) {
	query := `
SELECT "guild_id", "ticket_id"
FROM ticket_members
WHERE "user_id" = $1;`

	rows, err := db.Blacklist.Query(context.Background(), query, userId)
	must(err)

	for rows.Next() {
		var guildId uint64
		var ticketId int

		must(rows.Scan(&guildId, &ticketId))
		tickets = append(tickets, fmt.Sprintf("%d/%d", guildId, ticketId))
	}

	return
}

func getTickets(db *database.Database, userId uint64) (tickets []database.Ticket) {
	query := `
SELECT id, guild_id, channel_id, user_id, open, open_time, welcome_message_id, panel_id, has_transcript
FROM tickets
WHERE "user_id" = $1;`

	rows, err := db.Blacklist.Query(context.Background(), query, userId)
	must(err)

	for rows.Next() {
		var ticket database.Ticket
		must(rows.Scan(&ticket.Id, &ticket.GuildId, &ticket.ChannelId, &ticket.UserId, &ticket.Open, &ticket.OpenTime, &ticket.WelcomeMessageId, &ticket.PanelId, &ticket.HasTranscript))
		tickets = append(tickets, ticket)
	}

	return
}

func getPremiumActivatedFor(db *database.Database, userId uint64) (guilds []uint64) {
	query := `
SELECT guild_id
FROM used_keys
WHERE "activated_by" = $1;`

	rows, err := db.Blacklist.Query(context.Background(), query, userId)
	must(err)

	for rows.Next() {
		var guildId uint64
		must(rows.Scan(&guildId))
		guilds = append(guilds, guildId)
	}

	return
}

// cache data
func getCacheData(cache *cache.PgCache, userId uint64) map[string]interface{} {
	data := make(map[string]interface{})

	user, ok := cache.GetUser(userId)
	if ok {
		data["user"] = user
	} else {
		data["user"] = nil
	}

	rows, err := cache.Query(context.Background(), `SELECT guild_id, data FROM members WHERE "user_id" = $1;`, userId)
	must(err)

	memberData := make(map[string]interface{})
	for rows.Next() {
		var guildId uint64
		var raw string

		must(rows.Scan(&guildId, &raw))

		memberData[strconv.FormatUint(guildId, 10)] = json.RawMessage([]byte(raw))
	}

	data["member_data"] = memberData

	return data
}
