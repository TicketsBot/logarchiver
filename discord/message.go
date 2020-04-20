package discord

type Message struct {
	Author      User         `json:"author"`
	Content     string       `json:"content"`
	Attachments []Attachment `json:"attachments"`
}
