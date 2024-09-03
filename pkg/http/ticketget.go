package http

import (
	"errors"
	"github.com/TicketsBot/logarchiver/pkg/s3client"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"strconv"
)

func (s *Server) ticketGetHandler(ctx *gin.Context) {
	guild, err := strconv.ParseUint(ctx.Query("guild"), 10, 64)
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

	client, ok, err := s.getClientForObject(ctx, guild, id)
	if err != nil {
		s.Logger.Error("Failed to get client for object", zap.Error(err))
		ctx.JSON(500, gin.H{
			"message": err.Error(),
		})
		return
	}

	if !ok {
		ctx.JSON(404, gin.H{
			"message": "ticket not found",
		})
		return
	}

	data, err := client.GetTicket(ctx, guild, id)
	if err != nil {
		var statusCode int
		if errors.Is(err, s3client.ErrTicketNotFound) {
			statusCode = 404
		} else {
			statusCode = 500
		}

		ctx.JSON(statusCode, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx.Data(200, "application/octet-stream", data)
}
