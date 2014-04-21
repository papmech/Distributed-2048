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
)

const (
	ERROR_LOG bool = true
	DEBUG_LOG bool = true

	REGISTER_RETRY_INTERVAL = 500
)

var LOGV = util.NewLogger(DEBUG_LOG, "DEBUG", os.Stdout)
var LOGE = util.NewLogger(ERROR_LOG, "ERROR", os.Stderr)

type gameServer struct {
	id       uint32
	hostname string
	port     int
	hostport string

	libpaxos libpaxos.Libpaxos
}

// NewGameServer creates an instance of a Game Server. It does not return
// until it has successfully joined the cluster of game servers.
func NewGameServer(centralServerHostPort, hostname string, port int) (GameServer, error) {
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

	gs.id = reply.GameServerID
	gs.libpaxos, err = libpaxos.NewLibpaxos(reply.GameServerID, gs.hostport, reply.Servers)
	if err != nil {
		LOGE.Println("Could not start libpaxos")
		LOGE.Println(err)
		return nil, err
	}
	gs.libpaxos.DecidedHandler(gs.handleDecided)

	LOGV.Printf("GS node %d loaded libpaxos\n", reply.GameServerID)

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

func (gs *gameServer) TestAddVote(moves []paxosrpc.Move) {
	gs.libpaxos.Propose(moves)
}
