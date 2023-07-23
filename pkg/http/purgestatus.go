package http

import (
	"errors"
	"github.com/TicketsBot/logarchiver/internal"
	"github.com/gin-gonic/gin"
	"strconv"
)

func (s *Server) purgeStatusHandler(ctx *gin.Context) {
	guildId, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(400, gin.H{
			"success": false,
			"message": "missing guild ID",
		})
		return
	}

	operation, err := s.RemoveQueue.GetOperation(guildId)
	if err != nil {
		if errors.Is(err, internal.ErrOperationNotFound) {
			ctx.JSON(404, gin.H{
				"success": false,
				"message": "operation not found",
			})
			return
		} else {
			ctx.JSON(500, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	}

	ctx.JSON(200, operation)
}
