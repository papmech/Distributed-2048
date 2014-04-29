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
	"log"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

const (
	LOG_TO_FILE bool = false
	ERROR_LOG   bool = true
	DEBUG_LOG   bool = false

	REGISTER_RETRY_INTERVAL = 500
	CLIENT_UPDATE_INTERVAL  = 500
)

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
	doneCh       chan bool
	errCh        chan error
	numClients   int

	newMovesCh          chan *paxosrpc.ProposalValue
	totalNumGameServers int
	game2048            lib2048.Game2048
	stateBroadcastCh    chan *util.Game2048State
	clientMoveCh        chan *lib2048.Move

	LOGV *log.Logger
	LOGE *log.Logger
}

var movesOut *os.File

// NewGameServer creates an instance of a Game Server. It does not return
// until it has successfully joined the cluster of game servers.
func NewGameServer(centralServerHostPort, hostname string, port int, pattern string) (GameServer, error) {
	// Register with the central server
	c, err := rpc.DialHTTP("tcp", centralServerHostPort)
	if err != nil {
		fmt.Println("Could not connect to central server host port via RPC")
		fmt.Println(err)
		return nil, err
	}
	gshostport := fmt.Sprintf("%s:%d", hostname, port)
	args := &centralrpc.RegisterGameServerArgs{gshostport}
	var reply centralrpc.RegisterGameServerReply
	reply.Status = centralrpc.NotReady
	for reply.Status != centralrpc.OK {
		err = c.Call("CentralServer.RegisterGameServer", args, &reply)
		if err != nil {
			fmt.Println("Could not RPC call method CentralServer.RegisterGameServer")
			fmt.Println(err)
			return nil, err
		}
		if reply.Status == centralrpc.Full {
			return nil, errors.New("Could not register with central server, ring FULL")
		}
		time.Sleep(REGISTER_RETRY_INTERVAL * time.Millisecond)
	}
	// gs.LOGV.Printf("GS node %d finished registration with CS\n", reply.GameServerID)

	newlibpaxos, err := libpaxos.NewLibpaxos(reply.GameServerID, gshostport, reply.Servers)

	if err != nil {
		fmt.Println("Could not start libpaxos")
		fmt.Println(err)
		return nil, err
	}

	// Client related stuff
	clients := make(map[int]*client)
	addCh := make(chan *client)
	delCh := make(chan *client)
	//	sendAllCh := make(chan *Message)
	doneCh := make(chan bool)
	errCh := make(chan error)

	var vOut io.Writer
	var eOut io.Writer
	if LOG_TO_FILE {
		now := time.Now().String()
		var err error
		vOut, err = os.Create(fmt.Sprintf("%d_debug_%s.log", reply.GameServerID, now))
		if err != nil {
			panic(err)
		}
		eOut, err = os.Create(fmt.Sprintf("%d_error_%s.log", reply.GameServerID, now))
		if err != nil {
			panic(err)
		}
	} else {
		vOut = os.Stdout
		eOut = os.Stderr
	}

	movesOut, _ = os.Create(fmt.Sprintf("%d_moves_%s.log", reply.GameServerID, time.Now().String()))

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
		make(chan *paxosrpc.ProposalValue, 1000),
		len(reply.Servers),
		lib2048.NewGame2048(), // TODO: Use paxos to agree on game state
		make(chan *util.Game2048State, 1000),
		make(chan *lib2048.Move, 1000),
		util.NewLogger(DEBUG_LOG, "DEBUG", vOut),
		util.NewLogger(ERROR_LOG, "ERROR", eOut),
	}
	gs.libpaxos.DecidedHandler(gs.handleDecided)
	gs.LOGV.Printf("GS node %d loaded libpaxos\n", reply.GameServerID)

	go gs.ListenForClients()
	go gs.processMoves()
	// go gs.horseShit()
	go gs.clientTasker()
	go gs.clientMasterHandler()

	return gs, nil
}

func (gs *gameServer) GetLibpaxos() libpaxos.Libpaxos {
	return gs.libpaxos
}

func (gs *gameServer) clientListenRead(ws *websocket.Conn) {
	defer func() {
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
				gs.LOGE.Println(err)
			} else {
				var dir lib2048.Direction
				switch move.Direction {
				case 0:
					dir = lib2048.Up
				case 1:
					dir = lib2048.Right
				case 2:
					dir = lib2048.Down
				case 3:
					dir = lib2048.Left
				}
				gs.LOGV.Println("Received", dir, "from web client")
				move := lib2048.NewMove(dir)
				gs.clientMoveCh <- move
			}
		}
	}
}

func (gs *gameServer) clientMasterHandler() {
	ticker := time.NewTicker(500 * time.Millisecond) // send proposals every interval
	moves := make([]lib2048.Move, 0)
	for {
		select {
		case move := <-gs.clientMoveCh:
			moves = append(moves, *move)
		case <-ticker.C:
			if len(moves) > 0 {
				gs.libpaxos.Propose(&paxosrpc.ProposalValue{moves, *paxosrpc.NewGameData(gs.game2048)})
				moves = make([]lib2048.Move, 0)
			}
		}
	}
}

func (gs *gameServer) handleDecided(proposalValue *paxosrpc.ProposalValue) {
	gs.LOGV.Println(gs.id, "obtained decided proposal.")
	gs.newMovesCh <- proposalValue
}

func (gs *gameServer) processMoves() {
	sizeQueue := make([]int, 0)
	pendingMoves := make([]lib2048.Move, 0)
	currentBucketSize := 0
	for {
		select {
		case proposal := <-gs.newMovesCh:
			_, moves := proposal.Game, proposal.Moves

			// Reset the game state to the received state
			// game.CopyInto(gs.game2048)

			// Log the shit to file
			io.WriteString(movesOut, "-------------------------------------\n")
			io.WriteString(movesOut, util.MovesString(moves))
			io.WriteString(movesOut, "\n")

			pendingMoves = append(pendingMoves, moves...)
			if currentBucketSize == 0 {
				currentBucketSize = len(moves)
			} else {
				sizeQueue = append(sizeQueue, len(moves))
			}

			requiredMoves := currentBucketSize * gs.totalNumGameServers
			if len(pendingMoves) >= requiredMoves {
				// Find the majority of the first $requiredMoves moves
				dirVotes := make(map[lib2048.Direction]int)
				dirVotes[lib2048.Up] = 0
				dirVotes[lib2048.Down] = 0
				dirVotes[lib2048.Left] = 0
				dirVotes[lib2048.Right] = 0
				pendingMovesSubset := pendingMoves[:requiredMoves]
				pendingMoves = pendingMoves[requiredMoves:]
				for _, move := range pendingMovesSubset {
					dirVotes[move.Direction]++
				}

				gs.LOGV.Println("GAME SERVER", gs.id, "currentBucketSize", currentBucketSize)
				gs.LOGV.Println("GAME SERVER", gs.id, "pendingMovesSubset", util.MovesString(pendingMovesSubset))

				var majorityDir lib2048.Direction
				maxVotes := 0
				for dir, votes := range dirVotes {
					if votes > maxVotes {
						maxVotes = votes
						majorityDir = dir
					} else if votes == maxVotes && dir > majorityDir {
						majorityDir = dir
					}
				}

				gs.LOGV.Println("GAME SERVER", gs.id, "got majority direction:", majorityDir)

				// Update the 2048 state
				gs.game2048.MakeMove(majorityDir)

				state := gs.getWrappedState()

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
			gs.LOGV.Printf("GAME SERVER %d sending to client\n%s\n", gs.id, state.String())
			gs.clientsMutex.Lock()
			for _, c := range gs.clients {
				// gs.LOGV.Println("GOTS")
				err := websocket.Message.Send(c.conn, string(buf))
				if err != nil {
					gs.LOGE.Println(err)
				}
			}
			gs.clientsMutex.Unlock()
		}
	}
}

func (gs *gameServer) ListenForClients() {
	gs.LOGV.Println("Listening for connection from new clients")

	// websocket handler
	onConnected := func(ws *websocket.Conn) {
		gs.LOGV.Println("Client has connected")

		// client has been connected: add the client to the list
		gs.clientsMutex.Lock()
		c := &client{gs.numClients, ws}
		gs.clients[gs.numClients] = c
		id := gs.numClients

		// Remove from map when dead
		defer func() {
			gs.LOGV.Println("Lost connection to client", id)
			gs.clientsMutex.Lock()
			delete(gs.clients, id)
			gs.clientsMutex.Unlock()
		}()

		gs.numClients += 1
		gs.clientsMutex.Unlock()

		// Send it the state
		state := gs.getWrappedState()
		buf, _ := json.Marshal(*state)
		err := websocket.Message.Send(ws, string(buf))
		if err != nil {
			gs.LOGE.Println(err)
		}

		gs.clientListenRead(ws)
	}
	http.Handle(gs.pattern, websocket.Handler(onConnected))

	for {
		select {}
	}
}

func (gs *gameServer) TestAddVote(moves []lib2048.Move) {
	gs.libpaxos.Propose(&paxosrpc.ProposalValue{moves, *paxosrpc.NewGameData(gs.game2048)})
}

func (gs *gameServer) getWrappedState() *util.Game2048State {
	return &util.Game2048State{
		Won:   gs.game2048.IsGameWon(),
		Over:  gs.game2048.IsGameOver(),
		Grid:  gs.game2048.GetBoard(),
		Score: gs.game2048.GetScore(),
	}
}
