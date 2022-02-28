package v2

import (
	"fmt"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/guild"
	"github.com/rxdn/gdl/objects/user"
)

type Entities struct {
	Users    []User    `json:"users"`
	Channels []Channel `json:"channels"`
	Roles    []Role    `json:"roles"`
}

type User struct {
	Id            uint64             `json:"id,string"`
	Username      string             `json:"username"`
	Discriminator user.Discriminator `json:"discriminator"`
	Avatar        string             `json:"avatar"`
	Bot           bool               `json:"bot"`
}

func (u *User) AvatarUrl(size int) string {
	return fmt.Sprintf("https://cdn.discordapp.com/avatars/%d/%s.webp?size=%d", u.Id, u.Avatar, size)
}

type Channel struct {
	Id   uint64              `json:"id"`
	Name string              `json:"name"`
	Type channel.ChannelType `json:"type"`
}

type Role struct {
	Id     uint64 `json:"id"`
	Name   string `json:"name"`
	Colour uint32 `json:"color"`
}

func UserFromGdl(entity user.User) User {
	return User{
		Id:            entity.Id,
		Username:      entity.Username,
		Discriminator: entity.Discriminator,
		Avatar:        entity.Avatar.String(),
		Bot:           entity.Bot,
	}
}

func ChannelFromGdl(entity channel.Channel) Channel {
	return Channel{
		Id:   entity.Id,
		Name: entity.Name,
		Type: entity.Type,
	}
}

func RoleFromGdl(entity guild.Role) Role {
	return Role{
		Id:     entity.Id,
		Name:   entity.Name,
		Colour: uint32(entity.Color),
	}
}
