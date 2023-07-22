package http

import (
	"github.com/gin-gonic/gin"
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

	data, err := s.GetTicket(s.Config.Bucket, guild, id)
	if err != nil {
		var statusCode int
		if err == ErrTicketNotFound {
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
