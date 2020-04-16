package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/rxdn/gdl/objects/channel/message"
)

func EncodeHandler(ctx *gin.Context) {
	var messages []message.Message

	if err := ctx.BindJSON(&messages); err != nil {
		ctx.String(500, err.Error())
		return
	}

	ctx.HTML(200, "log.tmpl", gin.H{
		"ticketId": ctx.Query("id"),
		"messages": messages,
	})
}
