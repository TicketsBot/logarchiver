package http

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func (s *Server) ticketDeleteHandler(ctx *gin.Context) {
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

	if err := s.s3.DeleteTicket(ctx, guild, id); err != nil {
		ctx.JSON(500, gin.H{
			"message": err.Error(),
		})
	}

	ctx.Status(http.StatusNoContent)
}
