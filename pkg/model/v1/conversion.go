package v1

import (
	"github.com/TicketsBot/logarchiver/pkg/model"
	v22 "github.com/TicketsBot/logarchiver/pkg/model/v2"
	"github.com/rxdn/gdl/objects/channel/message"
)

func ConvertToV2(messages []message.Message) v22.Transcript {
	mappedMessages := make([]v22.Message, len(messages))
	users := make(map[uint64]v22.User)
	for i, message := range messages {
		mappedMessages[i] = v22.MessageFromGdl(message)
		users[message.Author.Id] = v22.UserFromGdl(message.Author)
	}

	return v22.Transcript{
		Version: model.V2,
		Entities: v22.Entities{
			Users: users,
		},
		Messages: mappedMessages,
	}
}
