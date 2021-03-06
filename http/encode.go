package http

import (
	"github.com/TicketsBot/logarchiver/discord"
	"github.com/gin-gonic/gin"
)

func encodeHandler(ctx *gin.Context) {
	var messages []discord.Message

	if err := ctx.BindJSON(&messages); err != nil {
		ctx.String(400, err.Error())
		return
	}

	ctx.HTML(200, "log.tmpl", gin.H{
		"title": ctx.Query("title"),
		"messages": messages,
	})
}
