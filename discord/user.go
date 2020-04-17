package discord

import "fmt"

type User struct {
	Id       uint64 `json:"id,string"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}

func (u *User) AvatarUrl(size int) string {
	return fmt.Sprintf("https://cdn.discordapp.com/avatars/%d/%s.webp?size=%d", u.Id, u.Avatar, size)
}
