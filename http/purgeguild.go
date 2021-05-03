package http

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
)

func (s *Server) purgeGuildHandler(ctx *gin.Context) {
	guildId, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(400, gin.H{
			"success": false,
			"message": "missing guild ID",
		})
		return
	}

	removeCh := make(chan string)
	go func() {
		for err := range s.client.RemoveObjects(s.Config.Bucket, removeCh) {
			s.Logger.Printf("error removing %s: %s\n", err.ObjectName, err.Err.Error())
		}
	}()

	go func() {
		doneCh := make(chan struct{})
		defer close(doneCh)

		objCh := s.client.ListObjectsV2(s.Config.Bucket, fmt.Sprintf("%d/", guildId), true, doneCh)

		for obj := range objCh {
			s.Logger.Printf("deleting %s\n", obj.Key)
			removeCh <- obj.Key
		}

		close(removeCh)
	}()

	ctx.JSON(200, gin.H{
		"success": true,
	})
}
