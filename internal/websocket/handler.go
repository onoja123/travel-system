package websocket

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

func ServeWS(hub *Hub, c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(400, gin.H{"error": "user_id required"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		UserID: userID,
	}

	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}
