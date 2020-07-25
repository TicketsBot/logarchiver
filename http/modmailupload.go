package http

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v6"
	"os"
)

func (s *Server) modmailUploadHandler(ctx *gin.Context) {
	body, err := ctx.GetRawData()
	if err != nil {
		ctx.JSON(400, gin.H{
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
	if _, err := s.client.PutObject(os.Getenv("S3_BUCKET"), name, bytes.NewReader(body), int64(len(body)), minio.PutObjectOptions{
		ContentType:     "application/octet-stream",
		ContentEncoding: "zstd",
	}); err != nil {
		ctx.JSON(500, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(200, gin.H{})
}
