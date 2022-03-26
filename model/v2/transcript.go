package v2

import (
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/channel/embed"
	"time"
)

type Transcript struct {
	Version  int       `json:"version"`
	Entities Entities  `json:"entities"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Id          uint64               `json:"id"`
	AuthorId    uint64               `json:"author"`
	Content     string               `json:"content"`
	Timestamp   time.Time            `json:"timestamp"`
	Embeds      []embed.Embed        `json:"embeds,omitempty"`
	Attachments []channel.Attachment `json:"attachments,omitempty"`
}
