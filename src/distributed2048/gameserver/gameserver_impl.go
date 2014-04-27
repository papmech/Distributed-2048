package gameserver

import (
	"code.google.com/p/go.net/websocket"
	"distributed2048/lib2048"
	"distributed2048/libpaxos"
	"distributed2048/rpc/centralrpc"
	"distributed2048/rpc/paxosrpc"
	"distributed2048/util"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

const (
	ERROR_LOG bool = true
	DEBUG_LOG bool = true

	REGISTER_RETRY_INTERVAL = 500
	CLIENT_UPDATE_INTERVAL  = 500
)

var LOGV = util.NewLogger(DEBUG_LOG, "DEBUG", os.Stdout)
var LOGE = util.NewLogger(ERROR_LOG, "ERROR", os.Stderr)

type client struct {
	id   int
	conn *websocket.Conn
}

type gameServer struct {
	id       uint32
	hostname string
	port     int
	hostport string

	libpaxos libpaxos.Libpaxos

	// Client-related stuff
	pattern      string
	clientsMutex sync.Mutex
	clients      map[int]*client
	addCh        chan *client
	delCh        chan *client
	doneCh     chan bool
	errCh      chan error
	numClients int

	newMovesCh          chan []paxosrpc.Move
	totalNumGameServers int
	game2048            lib2048.Game2048
	stateBroadcastCh    chan *util.Game2048State
	clientMoveCh        chan *paxosrpc.Move
}

// NewGameServer creates an instance of a Game Server. It does not return
// until it has successfully joined the cluster of game servers.
func NewGameServer(centralServerHostPort, hostname string, port int, pattern string) (GameServer, error) {
	// Register with the central server
	c, err := rpc.DialHTTP("tcp", centralServerHostPort)
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
		err = c.Call("CentralServer.RegisterGameServer", args, &reply)
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
		sync.Mutex{},
		clients,
		addCh,
		delCh,
		doneCh,
		errCh,
		0,
		make(chan []paxosrpc.Move, 1000),
		len(reply.Servers),
		lib2048.NewGame2048(), // TODO: Use paxos to agree on game state
		make(chan *util.Game2048State, 1000),
		make(chan *paxosrpc.Move, 1000),
	}
	gs.libpaxos.DecidedHandler(gs.handleDecided)
	LOGV.Printf("GS node %d loaded libpaxos\n", reply.GameServerID)

	go gs.ListenForClients()
	go gs.processMoves()
	// go gs.horseShit()
	go gs.clientTasker()
	go gs.clientMasterHandler()

	return gs, nil
}

// TODO: Remove this
// horseShit exists purely to test Paxos. At random
// intervals, it generates a random number of moves to be sent via Paxos.
func (gs *gameServer) horseShit() {
	time.Sleep(3 * time.Second)
	minMoves, maxMoves := 5, 20
	minSleep, maxSleep := 100, 400

	for {
		time.Sleep(time.Duration(rand.Intn(maxSleep-minSleep)+minSleep) * time.Millisecond)
		moves := make([]paxosrpc.Move, 0)
		numMoves := rand.Intn(maxMoves - minMoves)
		for i := 0; i < numMoves; i++ {
			moves = append(moves, *util.RandomMove())
		}
		gs.libpaxos.Propose(moves)
	}
}

func (gs *gameServer) handleDecided(slotNumber uint32, moves []paxosrpc.Move) {
	LOGV.Println("GAME SERVER", gs.id, "got slot", slotNumber)

	gs.newMovesCh <- moves

	// TODO remove the slotNumber since only libpaxos needs it

	// Send the move to the 2048 go routine, which will update itself
}

func (gs *gameServer) clientMasterHandler() {
	ticker := time.NewTicker(500 * time.Millisecond) // send proposals every interval
	moves := make([]paxosrpc.Move, 0)
	for {
		select {
		case move := <-gs.clientMoveCh:
			moves = append(moves, *move)
		case <-ticker.C:
			if len(moves) > 0 {
				gs.libpaxos.Propose(moves)
				moves = make([]paxosrpc.Move, 0)
			}
		}
	}
}

func (gs *gameServer) clientListenRead(ws *websocket.Conn) {
	defer func() {
		LOGV.Println("I'm hauling ass")
		ws.Close()
	}()

	for {
		select {
		default:
			var move util.ClientMove
			err := websocket.JSON.Receive(ws, &move)
			if err == io.EOF {
				return
				// EOF!
			} else if err != nil {
				LOGE.Println(err)
			} else {
				var dir paxosrpc.Direction
				switch move.Direction {
				case 0:
					dir = paxosrpc.Up
				case 1:
					dir = paxosrpc.Right
				case 2:
					dir = paxosrpc.Down
				case 3:
					dir = paxosrpc.Left
				}
				LOGV.Println("Received", dir, "from web client")
				move := paxosrpc.NewMove(dir)
				gs.clientMoveCh <- move
			}
		}
	}
}

func (gs *gameServer) processMoves() {
	sizeQueue := make([]int, 0)
	pendingMoves := make([]paxosrpc.Move, 0)
	currentBucketSize := 0
	for {
		select {
		case moves := <-gs.newMovesCh:
			pendingMoves = append(pendingMoves, moves...)
			if currentBucketSize == 0 {
				currentBucketSize = len(moves)
			} else {
				sizeQueue = append(sizeQueue, len(moves))
			}

			requiredMoves := currentBucketSize * gs.totalNumGameServers
			if len(pendingMoves) >= requiredMoves {
				// Find the majority of the first $requiredMoves moves
				dirVotes := make(map[paxosrpc.Direction]int)
				dirVotes[paxosrpc.Up] = 0
				dirVotes[paxosrpc.Down] = 0
				dirVotes[paxosrpc.Left] = 0
				dirVotes[paxosrpc.Right] = 0
				pendingMovesSubset := pendingMoves[:requiredMoves]
				pendingMoves = pendingMoves[requiredMoves:]
				for _, move := range pendingMovesSubset {
					dirVotes[move.Direction]++
				}

				LOGV.Println("GAME SERVER", gs.id, "currentBucketSize", currentBucketSize)
				LOGV.Println("GAME SERVER", gs.id, "pendingMovesSubset", util.MovesString(pendingMovesSubset))

				var majorityDir paxosrpc.Direction
				maxVotes := 0
				for dir, votes := range dirVotes {
					if votes > maxVotes {
						maxVotes = votes
						majorityDir = dir
					} else if votes == maxVotes && dir > majorityDir {
						majorityDir = dir
					}
				}

				LOGV.Println("GAME SERVER", gs.id, "got majority direction:", majorityDir)

				// Update the 2048 state
				gs.game2048.MakeMove(majorityDir)

				state := &util.Game2048State{
					Won:   gs.game2048.IsGameWon(),
					Over:  gs.game2048.IsGameOver(),
					Grid:  gs.game2048.GetBoard(),
					Score: gs.game2048.GetScore(),
				}

				gs.stateBroadcastCh <- state

				// Update the bucket size
				if len(sizeQueue) > 0 {
					currentBucketSize = sizeQueue[0]
					sizeQueue = sizeQueue[1:]
				}
			}

		}
	}
}

func (gs *gameServer) clientTasker() {
	for {
		select {
		case state := <-gs.stateBroadcastCh:
			buf, _ := json.Marshal(*state)
			LOGV.Println("GAME SERVER", gs.id, "sendin to client:", string(buf))
			gs.clientsMutex.Lock()
			for _, c := range gs.clients {
				LOGV.Println("GOTS")
				err := websocket.Message.Send(c.conn, string(buf))
				if err != nil {
					LOGE.Println(err)
				}
			}
			gs.clientsMutex.Unlock()
		}
	}
}

func (gs *gameServer) ListenForClients() {
	LOGV.Println("Listening for connection from new clients")

	// websocket handler
	onConnected := func(ws *websocket.Conn) {
		LOGV.Println("Client has connected")
		// client has been connected: add the client to the list
		gs.clientsMutex.Lock()
		c := &client{gs.numClients, ws}
		gs.clients[gs.numClients] = c
		gs.numClients += 1
		gs.clientsMutex.Unlock()
		state := &util.Game2048State{
			Won:   gs.game2048.IsGameWon(),
			Over:  gs.game2048.IsGameOver(),
			Grid:  gs.game2048.GetBoard(),
			Score: gs.game2048.GetScore(),
		}
		gs.stateBroadcastCh <- state
		gs.clientListenRead(ws)
	}
	http.Handle(gs.pattern, websocket.Handler(onConnected))

	for {
		select {
		}
	}
}

func (gs *gameServer) TestAddVote(moves []paxosrpc.Move) {
	gs.libpaxos.Propose(moves)
}

func (gs *gameServer) handleNewMoveList() {
	for {
		select {
		case moves := <-gs.newMovesCh:
			LOGV.Println(moves)
		}
	}
}
