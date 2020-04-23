package http

import (
	"fmt"
	"github.com/TicketsBot/logarchiver"
	"github.com/TicketsBot/logarchiver/config"
	"github.com/gin-gonic/gin"
	"sort"
	"strconv"
)

func (s *Server) modmailListHandler(ctx *gin.Context) {
	guildId, err := strconv.ParseUint(ctx.Query("guild"), 10, 64)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"success": false,
			"error":   "Missing guild ID",
		})
		return
	}

	done := make(chan struct{})
	defer close(done)

	prefix := fmt.Sprintf("%d/modmail/", guildId)

	objects := make([]logarchiver.StoredObject, 0)
	for object := range s.client.ListObjectsV2WithMetadata(config.Conf.S3.Bucket, prefix, true, done) {
		if object.Err != nil {
			ctx.AbortWithStatusJSON(500, gin.H{
				"success": false,
				"error":   object.Err.Error(),
			})
			return
		}

		if object.Key == prefix {
			continue
		}

		objects = append(objects, logarchiver.StoredObject{
			Key:          object.Key,
			LastModified: object.LastModified,
		})
	}

	sort.Slice(objects, func(i, j int) bool {
		return objects[i].LastModified.Before(objects[j].LastModified)
	})

	ctx.JSON(200, objects)
}
