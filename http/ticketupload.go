package http

import (
	"bytes"
	"fmt"
	"github.com/TicketsBot/logarchiver/config"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v6"
)

func (s *Server) ticketUploadHandler(ctx *gin.Context) {
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

	var id string
	id, ok = ctx.GetQuery("id")
	if !ok {
		ctx.JSON(400, gin.H{
			"message": "missing ticket ID",
		})
		return
	}

	var freePrefix string
	if _, premium := ctx.GetQuery("premium"); !premium {
		freePrefix = "free-"
	}

	name := fmt.Sprintf("%s/%s%s", guild, freePrefix, id)

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
