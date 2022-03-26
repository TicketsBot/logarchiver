package v1

import (
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/channel/embed"
	"time"
)

type Message struct {
	Author      User                 `json:"author"`
	Content     string               `json:"content"`
	Timestamp   time.Time            `json:"timestamp"`
	Embeds      []embed.Embed        `json:"embeds,omitempty"`
	Attachments []channel.Attachment `json:"attachments,omitempty"`
}
