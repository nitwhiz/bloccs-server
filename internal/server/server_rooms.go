package server

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func postRoomsHandler(s *Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		room := s.CreateRoom()

		c.JSON(http.StatusCreated, gin.H{
			"roomId": room.Id,
		})
	}
}

func postRoomStartHandler(s *Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		roomId := c.Param("roomId")

		if roomId == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "room id is empty",
			})

			return
		}

		room := s.GetRoom(roomId)

		if room == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "room not found",
			})

			return
		}

		room.Start()

		c.Status(http.StatusNoContent)
	}
}

func getRoomsHandler(s *Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		roomId := c.Param("roomId")

		if roomId == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "room id is empty",
			})

			return
		}

		room := s.GetRoom(roomId)

		if room == nil {
			c.Status(http.StatusNotFound)
		} else {
			c.Status(http.StatusNoContent)
		}
	}
}

func getRoomSocketHandler(s *Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		roomId := c.Param("roomId")

		if _, ok := s.rooms[roomId]; !ok {
			c.AbortWithStatus(http.StatusNotFound)
		}

		if err := s.connect(roomId, c.Writer, c.Request); err != nil {
			c.Abort()
		}
	}
}
