package http

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

func (s *Server) ticketDeleteHandler(ctx *gin.Context) {
	guildId, err := strconv.ParseUint(ctx.Query("guild"), 10, 64)
	if err != nil {
		ctx.JSON(400, gin.H{
			"message": "missing guild ID",
		})
		return
	}

	id, err := strconv.Atoi(ctx.Query("id"))
	if err != nil {
		ctx.JSON(400, gin.H{
			"message": "missing ticket ID",
		})
		return
	}

	logger := s.Logger.With(zap.Uint64("guild", guildId), zap.Int("ticket", id))

	// Find bucket
	client, ok, err := s.getClientForObject(ctx, guildId, id)
	if err != nil {
		logger.Error("Failed to get client for object", zap.Error(err))
		ctx.JSON(500, gin.H{
			"message": err.Error(),
		})
		return
	}

	if !ok {
		logger.Warn("Ticket not found")
		ctx.JSON(404, gin.H{
			"message": "ticket not found",
		})
		return
	}

	if err := client.DeleteTicket(ctx, guildId, id); err != nil {
		logger.Error("Failed to delete ticket", zap.Error(err))
		ctx.JSON(500, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx.Status(http.StatusNoContent)
}
