package discord

import "time"

type Message struct {
	Author      User         `json:"author"`
	Content     string       `json:"content"`
	Timestamp   time.Time    `json:"timestamp"`
	Attachments []Attachment `json:"attachments"`
}
