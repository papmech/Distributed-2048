package gameserver

import (
	"distributed2048/libpaxos"
	"distributed2048/rpc/centralrpc"
	"distributed2048/util"
	"errors"
	"fmt"
	"net/rpc"
	"os"
	"time"
	"net/http"
	"code.google.com/p/go.net/websocket"
)

const (
	ERROR_LOG bool = true
	DEBUG_LOG bool = true

	REGISTER_RETRY_INTERVAL = 500
)

var LOGV = util.NewLogger(DEBUG_LOG, "DEBUG", os.Stdout)
var LOGE = util.NewLogger(ERROR_LOG, "ERROR", os.Stderr)

type Client struct {
	id int

}

type gameServer struct {
	hostname string
	port     int
	hostport string

	libpaxos libpaxos.Libpaxos
	pattern string
	clients map[int]*Client
}

// NewGameServer creates an instance of a Game Server. It does not return
// until it has successfully joined the cluster of game servers.
func NewGameServer(centralServerHostPort, hostname string, port int, pattern string) (GameServer, error) {
	gs := &gameServer{
		hostname: hostname,
		port:     port,
		hostport: fmt.Sprintf("%s:%d", hostname, port),
	}

	// Register with the central server
	client, err := rpc.DialHTTP("tcp", centralServerHostPort)
	if err != nil {
		LOGE.Println("Could not connect to central server host port via RPC")
		LOGE.Println(err)
		return nil, err
	}
	args := &centralrpc.RegisterGameServerArgs{gs.hostport}
	var reply centralrpc.RegisterGameServerReply
	reply.Status = centralrpc.NotReady
	for reply.Status != centralrpc.OK {
		err = client.Call("CentralServer.RegisterGameServer", args, &reply)
		if err != nil {
			LOGE.Println("Could not RPC call method CentralServer.RegisterGameServer")
			LOGE.Println(err)
			return nil, err
		}
		if reply.Status == centralrpc.Full {
			return nil, errors.New("Could not register with central server, ring FULL")
		}
		time.Sleep(REGISTER_RETRY_INTERVAL * time.Millisecond)
	}
	LOGV.Printf("GS node %d finished registration with CS\n", reply.GameServerID)

	gs.libpaxos, err = libpaxos.NewLibpaxos(reply.GameServerID, gs.hostport, reply.Servers)
	if err != nil {
		LOGE.Println("Could not start libpaxos")
		LOGE.Println(err)
		return nil, err
	}

	LOGV.Printf("GS node %d loaded libpaxos\n", reply.GameServerID)

	clients := make(map[int]*Client)
	return &Server {
		pattern,
		clients,
	}

	return gs, nil
}

func (gs *gameServer) DoVote() {

}

func (gs *gameServer) AddVote() {

}

func (gs *gameServer) SetVoteResult() {
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
