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

	// Client-related stuff
	pattern string
	clients map[int]*Client
	addCh     chan *Client
	delCh     chan *Client
//	sendAllCh chan *Message
	doneCh    chan bool
	errCh     chan error
}

// NewGameServer creates an instance of a Game Server. It does not return
// until it has successfully joined the cluster of game servers.
func NewGameServer(centralServerHostPort, hostname string, port int, pattern string) (GameServer, error) {
	// Register with the central server
	client, err := rpc.DialHTTP("tcp", centralServerHostPort)
	if err != nil {
		LOGE.Println("Could not connect to central server host port via RPC")
		LOGE.Println(err)
		return nil, err
	}
	gshostport := fmt.Sprintf("%s:%d", hostname, port)
	args := &centralrpc.RegisterGameServerArgs{gshostport}
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

	libpaxos, err := libpaxos.NewLibpaxos(reply.GameServerID, gshostport, reply.Servers)
	if err != nil {
		LOGE.Println("Could not start libpaxos")
		LOGE.Println(err)
		return nil, err
	}

	LOGV.Printf("GS node %d loaded libpaxos\n", reply.GameServerID)

	// Client related stuff
	clients := make(map[int]*Client)
	addCh := make(chan *Client)
	delCh := make(chan *Client)
//	sendAllCh := make(chan *Message)
	doneCh := make(chan bool)
	errCh := make(chan error)

	gs := &gameServer{
		hostname,
		port,
		gshostport,
		libpaxos,
		pattern,
//		messages,
		clients,
		addCh,
		delCh,
//		sendAllCh,
		doneCh,
		errCh,

	}

	go gs.ListenForClients()
	return gs, nil
}

func (gs *gameServer) DoVote() {

}

func (gs *gameServer) AddVote() {

}

func (gs *gameServer) SetVoteResult() {
}

func (gs *gameServer) ListenForClients() {
	LOGV.Println("Listening for connection from new clients")

	// websocket handler
	onConnected := func(ws *websocket.Conn) {
		LOGV.Print("Client has connected")
//		defer func() {
//			err := ws.Close()
//			if err != nil {
//				gs.errCh <- err
//			}
//		}()

		//		client := NewClient(ws, s)
	//		s.Add(client)
		//		client.Listen()
	}
	http.Handle(gs.pattern, websocket.Handler(onConnected))
	LOGV.Println("Created handler")

	for {
		select {
			// Add a new client
//		case c := <-gs.addCh:
//			LOGV.Println("added new client")

			// Delete a client
//		case c := <-gs.delCh:
			// Broadcast a message to all clients
//		case msg := <-gs.sendAllCh:
			// Error channel
//		case err := <-gs.errCh:

		}
	}
}
