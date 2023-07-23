package http

import (
	"errors"
	"fmt"
	"github.com/TicketsBot/logarchiver/internal"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
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

	if err := s.RemoveQueue.StartOperation(guildId); err != nil {
		if errors.Is(err, internal.ErrOperationInProgress) {
			ctx.JSON(400, gin.H{
				"success": false,
				"message": "operation already in progress",
			})
		} else {
			ctx.JSON(500, gin.H{
				"success": false,
				"message": err.Error(),
			})
		}

		return
	}

	removeCh := make(chan string)
	go func() {
		var errCount int
		for err := range s.client.RemoveObjects(s.Config.Bucket, removeCh) {
			s.RemoveQueue.AddError(guildId, err.ObjectName, err.Err)

			s.Logger.Error(
				"Failed to remove object",
				zap.Error(err.Err),
				zap.String("object", err.ObjectName),
				zap.Uint64("guild", guildId),
			)

			errCount++
		}

		if errCount > 0 {
			s.RemoveQueue.SetStatus(guildId, internal.StatusFailed)
			s.Logger.Warn(
				"Remove operation completed with error(s)",
				zap.Int("error_count", errCount),
				zap.Uint64("guild", guildId),
			)
		} else {
			s.RemoveQueue.SetStatus(guildId, internal.StatusComplete)
			s.Logger.Info("Remove operation completed successfully", zap.Uint64("guild", guildId))
		}
	}()

	go func() {
		doneCh := make(chan struct{})
		defer close(doneCh)

		objCh := s.client.ListObjectsV2(s.Config.Bucket, fmt.Sprintf("%d/", guildId), true, doneCh)

		for obj := range objCh {
			s.Logger.Info(
				"Found object to remove",
				zap.String("object", obj.Key),
				zap.Uint64("guild", guildId),
			)

			s.RemoveQueue.AddRemovedObject(guildId, obj.Key)
			removeCh <- obj.Key
		}

		close(removeCh)
	}()

	ctx.JSON(http.StatusAccepted, gin.H{
		"success": true,
	})
}
