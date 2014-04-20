package gameserver

import (
	"os"
	"distributed2048/util"
	"net/http"
	"code.google.com/p/go.net/websocket"
)

const (
	ERROR_LOG bool = true
	DEBUG_LOG bool = true
)

var LOGV = util.NewLogger(DEBUG_LOG, "DEBUG", os.Stdout)
var LOGE = util.NewLogger(ERROR_LOG, "ERROR", os.Stderr)

type Client struct {
	id int

}

type GameServer struct {
	pattern string
	clients map[int]*Client

}

func NewServer(pattern string) *GameServer {
	clients := make(map[int]*Client)
	return &Server {
		pattern,
		clients,
	}
}

func (s *GameServer) ListenForClients() {
	LOGV.Println("Listening for connection from new clients")

	// websocket handler
	onConnected := func(ws *websocket.Conn) {
		defer func() {
			err := ws.Close()
			if err != nil {
				s.errCh <- err
			}
		}()

//		client := NewClient(ws, s)
		s.Add(client)
//		client.Listen()
	}
	http.Handle(s.pattern, websocket.Handler(onConnected))
	LOGV.Println("Created handler")

	for {
		select {
		// Add a new client
		case c := <-s.addCh:
			LOGV.Println("added new client")

		// Delete a client
		case c := <-s.delCh:
		// Broadcast a message to all clients
		case msg := <-s.sendAllCh:
		// Error channel
		case err := <-s.errCh:

		}
	}


}
