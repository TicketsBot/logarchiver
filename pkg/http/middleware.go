package http

import "github.com/gin-gonic/gin"

func (s *Server) middlewareAuthAdmin(ctx *gin.Context) {
	if len(s.Config.AdminAuthToken) == 0 {
		s.Logger.Error("Admin authentication token not set")
		ctx.JSON(500, gin.H{
			"message": "misconfigured server",
		})
		ctx.Abort()
		return
	}

	if ctx.GetHeader("Authorization") != s.Config.AdminAuthToken {
		ctx.JSON(401, gin.H{
			"message": "unauthorized",
		})
		ctx.Abort()
		return
	}

	ctx.Next()
}
