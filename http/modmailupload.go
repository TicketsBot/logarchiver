package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/TicketsBot/logarchiver/config"
	"github.com/TicketsBot/logarchiver/discord"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v6"
)

func (s *Server) modmailUploadHandler(ctx *gin.Context) {
	var messages []discord.Message

	if err := ctx.BindJSON(&messages); err != nil {
		ctx.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	// re-marshal to our own format
	encoded, err := json.Marshal(&messages)
	if err != nil {
		ctx.JSON(500, gin.H{
			"message": err.Error(),
		})
		return
	}

	guild, ok := ctx.GetQuery("guild")
	if !ok {
		ctx.JSON(400, gin.H{
			"message": "missing guild ID",
		})
		return
	}

	var uuid string
	uuid, ok = ctx.GetQuery("uuid")
	if !ok {
		ctx.JSON(400, gin.H{
			"message": "missing ticket UUID",
		})
		return
	}

	var freePrefix string
	if _, premium := ctx.GetQuery("premium"); !premium {
		freePrefix = "free-"
	}

	name := fmt.Sprintf("%s/modmail/%s%s", guild, freePrefix, uuid)

	// DigitalOcean does not support RetailUntilDate
	if _, err := s.client.PutObject(config.Conf.S3.Bucket, name, bytes.NewReader(encoded), int64(len(encoded)), minio.PutObjectOptions{
		ContentType:     "application/json",
	}); err != nil {
		ctx.JSON(500, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(200, gin.H{})
}
