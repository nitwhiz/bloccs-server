package server

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func getPingHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"online": true,
		})
	}
}
