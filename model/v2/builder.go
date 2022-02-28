package v2

import (
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/guild"
	"github.com/rxdn/gdl/objects/user"
)

func NewTranscript(
	messages []message.Message,
	userRetriever func(userIds []uint64) []user.User,
	channelRetriever func(channelIds []uint64) []channel.Channel,
	roleRetriever func(roleIds []uint64) guild.Role,
) Transcript {
	reduced := make([]Message, len(messages))

	users := make(map[uint64]User)
	channels := make(map[uint64]Channel)
	roles := make(map[uint64]Role)

	for i, message := range messages {
		reduced[i] = Message{
			Id:          message.Id,
			AuthorId:    message.Author.Id,
			Content:     message.Content,
			Timestamp:   message.Timestamp,
			Embeds:      message.Embeds,
			Attachments: message.Attachments,
		}

		users[message.Author.Id] = UserFromGdl(message.Author)

		// TODO: Regex for user + channel + role mentions
	}

	entities := Entities{}

	{
		i := 0
		for id := range users {
			userIds[i] = id
			i++
		}
	}

	{
		i := 0
		for id := range channels {
			channelIds[i] = id
			i++
		}
	}

	{
		i := 0
		for id := range roles {
			roleIds[i] = id
			i++
		}
	}

	return Transcript{
		Version:  2,
		Entities: entities,
		Messages: reduced,
	}
}
