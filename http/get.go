package http

import (
	"bytes"
	"fmt"
	"github.com/TicketsBot/logarchiver/config"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v6"
)

func (s *Server) getHandler(ctx *gin.Context) {
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

	// try reading with free name
	reader, err := s.client.GetObject(config.Conf.S3.Bucket, fmt.Sprintf("%s/free-%s", guild, id), minio.GetObjectOptions{})
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
			ctx.Data(200, "application/json", buff.Bytes())
		}
	}

	// else, we should check the premium object
	reader, err = s.client.GetObject(config.Conf.S3.Bucket, fmt.Sprintf("%s/%s", guild, id), minio.GetObjectOptions{})
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
			ctx.Data(200, "application/json", buff.Bytes())
			return
		}
	}

	// if we didn't find the object, then it doesn't exist in our store
	ctx.JSON(404, gin.H{
		"message": "object not found",
	})
}
