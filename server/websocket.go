package server

import (
	"net/http"

	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/gin-gonic/gin"

	ws "github.com/chennqqi/godnslog/internal/websocket"
)

var wsHub = ws.NewHub()

// initWS starts the WebSocket hub event loop.
func initWS() {
	go wsHub.Run()
	logrus.Info("[websocket] WebSocket hub started")
}

// wsHandler upgrades HTTP to WebSocket and registers the client.
func wsHandler(c *gin.Context) {
	upgrader := gorillawebsocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins (OAST tool, acceptable)
		},
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logrus.Warnf("[websocket] upgrade failed: %v", err)
		return
	}

	client := ws.NewClient(wsHub, conn)
	wsHub.Register(client)
	go client.WritePump()
	go client.ReadPump()
}
