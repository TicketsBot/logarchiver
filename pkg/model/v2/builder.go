package v2

import (
	"github.com/TicketsBot/logarchiver/pkg/model"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/guild"
	"github.com/rxdn/gdl/objects/user"
	"regexp"
	"strconv"
)

var (
	userRegex    = regexp.MustCompile(`<@!?(\d{16,20})>`)
	roleRegex    = regexp.MustCompile(`<@&(\d{16,20})>`)
	channelRegex = regexp.MustCompile(`<#(\d{16,20})>`)
)

func NewTranscript(
	messages []message.Message,
	userRetriever func(userIds []uint64) []user.User,
	channelRetriever func(channelIds []uint64) []channel.Channel,
	roleRetriever func(roleIds []uint64) []guild.Role,
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

		// Match snowflakes in message
		userMatches := userRegex.FindAllStringSubmatch(message.Content, -1)
		roleMatches := roleRegex.FindAllStringSubmatch(message.Content, -1)
		channelMatches := channelRegex.FindAllStringSubmatch(message.Content, -1)

		for _, user := range userRetriever(extractSnowflakes(userMatches)) {
			users[user.Id] = UserFromGdl(user)
		}

		for _, role := range roleRetriever(extractSnowflakes(roleMatches)) {
			roles[role.Id] = RoleFromGdl(role)
		}

		for _, channel := range channelRetriever(extractSnowflakes(channelMatches)) {
			channels[channel.Id] = ChannelFromGdl(channel)
		}
	}

	entities := Entities{
		Users:    users,
		Channels: channels,
		Roles:    roles,
	}

	return Transcript{
		Version:  model.V2,
		Entities: entities,
		Messages: reduced,
	}
}

func extractSnowflakes(matches [][]string) []uint64 {
	snowflakes := make([]uint64, len(matches))

	for _, match := range matches {
		snowflake, _ := strconv.ParseUint(match[1], 10, 64)
		snowflakes = append(snowflakes, snowflake)
	}

	return snowflakes
}

func NoopRetriever[T any](snowflakes []uint64) []T {
	return nil
}
