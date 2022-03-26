package v1

import (
	"github.com/TicketsBot/logarchiver/model"
	v2 "github.com/TicketsBot/logarchiver/model/v2"
	"github.com/rxdn/gdl/objects/channel/message"
)

func ConvertToV2(messages []message.Message) v2.Transcript {
	mappedMessages := make([]v2.Message, len(messages))
	users := make(map[uint64]v2.User)
	for i, message := range messages {
		mappedMessages[i] = v2.MessageFromGdl(message)
		users[message.Author.Id] = v2.UserFromGdl(message.Author)
	}

	return v2.Transcript{
		Version: model.V2,
		Entities: v2.Entities{
			Users: users,
		},
		Messages: mappedMessages,
	}
}
