package http

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v6"
	"os"
)

func (s *Server) modmailGetHandler(ctx *gin.Context) {
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

	// try reading with free name
	reader, err := s.client.GetObject(os.Getenv("S3_BUCKET"), fmt.Sprintf("%s/modmail/free-%s", guild, uuid), minio.GetObjectOptions{})
	if err != nil {
		ctx.JSON(500, gin.H{
			"message": err.Error(),
		})
		return
	}

	// if we found the free object, we can return it
	if reader != nil {
		defer reader.Close()

		var buff bytes.Buffer
		_, err = buff.ReadFrom(reader)
		if err != nil {
			if err.Error() != "The specified key does not exist." {
				ctx.JSON(500, gin.H{
					"message": err.Error(),
				})
				return
			}
		} else {
			ctx.Data(200, "application/octet-stream", buff.Bytes())
			return
		}
	}

	// else, we should check the premium object
	reader, err = s.client.GetObject(os.Getenv("S3_BUCKET"), fmt.Sprintf("%s/modmail/%s", guild, uuid), minio.GetObjectOptions{})
	if err != nil {
		ctx.JSON(500, gin.H{
			"message": err.Error(),
		})
		return
	}

	if reader != nil {
		defer reader.Close()

		var buff bytes.Buffer
		_, err = buff.ReadFrom(reader)
		if err != nil {
			if err.Error() != "The specified key does not exist." {
				ctx.JSON(500, gin.H{
					"message": err.Error(),
				})
				return
			}
		} else {
			ctx.Data(200, "application/octet-stream", buff.Bytes())
			return
		}
	}

	// if we didn't find the object, then it doesn't exist in our store
	ctx.JSON(404, gin.H{
		"message": "object not found",
	})
}
