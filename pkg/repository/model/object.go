package model

import "github.com/google/uuid"

type Object struct {
	GuildId  uint64    `json:"guild_id,string"`
	TicketId int       `json:"ticket_id"`
	BucketId uuid.UUID `json:"bucket_id"`
}
