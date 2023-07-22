package v2

import (
	"github.com/TicketsBot/logarchiver/pkg/model"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/channel/message"
	"time"
)

type Transcript struct {
	Version  model.Version `json:"version"`
	Entities Entities      `json:"entities"`
	Messages []Message     `json:"messages"`
}

type Message struct {
	Id          uint64               `json:"id"`
	AuthorId    uint64               `json:"author"`
	Content     string               `json:"content"`
	Timestamp   time.Time            `json:"timestamp"`
	Embeds      []embed.Embed        `json:"embeds,omitempty"`
	Attachments []channel.Attachment `json:"attachments,omitempty"`
}

func MessageFromGdl(message message.Message) Message {
	return Message{
		Id:          message.Id,
		AuthorId:    message.Author.Id,
		Content:     message.Content,
		Timestamp:   message.Timestamp,
		Embeds:      message.Embeds,
		Attachments: message.Attachments,
	}
}
