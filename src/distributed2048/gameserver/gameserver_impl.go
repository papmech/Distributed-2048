package gameserver

import (
	"distributed2048/libpaxos"
	"distributed2048/rpc/centralrpc"
	"distributed2048/rpc/paxosrpc"
	"distributed2048/util"
	"errors"
	"fmt"
	"math/rand"
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

type client struct {
	id int
	conn *websocket.Conn
}

type gameServer struct {
	id       uint32
	hostname string
	port     int
	hostport string

	libpaxos libpaxos.Libpaxos

	// Client-related stuff
	pattern string
	clients map[int]*client
	addCh     chan *client
	delCh     chan *client
//	sendAllCh chan *Message
	doneCh    chan bool
	errCh     chan error
	numClients int
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

	newlibpaxos, err := libpaxos.NewLibpaxos(reply.GameServerID, gshostport, reply.Servers)

	if err != nil {
		LOGE.Println("Could not start libpaxos")
		LOGE.Println(err)
		return nil, err
	}

	// Client related stuff
	clients := make(map[int]*client)
	addCh := make(chan *client)
	delCh := make(chan *client)
//	sendAllCh := make(chan *Message)
	doneCh := make(chan bool)
	errCh := make(chan error)

	gs := &gameServer{
		reply.GameServerID,
		hostname,
		port,
		gshostport,
		newlibpaxos,
		pattern,
//		messages,
		clients,
		addCh,
		delCh,
//		sendAllCh,
		doneCh,
		errCh,
		0,
	}
	gs.libpaxos.DecidedHandler(gs.handleDecided)
	LOGV.Printf("GS node %d loaded libpaxos\n", reply.GameServerID)

	go gs.ListenForClients()
	go gs.horseShit()

	return gs, nil
}

func (gs *gameServer) horseShit() {
	time.Sleep(3 * time.Second)
	for {
		if gs.id == 0 {
			num := time.Duration(rand.Int() % 50)
			time.Sleep(num * time.Millisecond)
			// LOGV.Println("Horse shit starting on", gs.id)
			moves := []paxosrpc.Move{
				*paxosrpc.NewMove(paxosrpc.Up),
				*paxosrpc.NewMove(paxosrpc.Up),
			}
			gs.libpaxos.Propose(moves)
		} else if gs.id == 1 {
			num := time.Duration(rand.Int() % 50)
			time.Sleep(num * time.Millisecond)
			// LOGV.Println("Horse shit starting on", gs.id)
			moves := []paxosrpc.Move{
				*paxosrpc.NewMove(paxosrpc.Down),
				*paxosrpc.NewMove(paxosrpc.Left),
			}
			gs.libpaxos.Propose(moves)
		} else {
			select {}
		}
	}
}

func (gs *gameServer) handleDecided(moves []paxosrpc.Move) {
	// LOGV.Println("Holy shit! Paxos quorum round has complete, decided moves:")
	// LOGV.Println(util.MovesString(moves))
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
		LOGV.Println("Client has connected")
		// client has been connected: add the client to the list
		client := &client(gs.numClients, ws)
		gs.clients[gs.numClients] = client
		gs.numClients += 1

		for {
			var in []byte
			if err := websocket.Message.Receive(ws, &in); err != nil {
				LOGE.Println("Error when receiving message from client")
			}
			fmt.Printf("Received: %s\n", string(in))
			websocket.Message.Send(ws, in)
		}
//		for {
//			LOGV.Println("I r t3h spammingz")
//			time.Sleep(1000)
//		}
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

func (gs *gameServer) TestAddVote(moves []paxosrpc.Move) {
	gs.libpaxos.Propose(moves)
}
