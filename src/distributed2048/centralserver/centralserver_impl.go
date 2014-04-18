package centralserver

import (
	"distributed2048/rpc/centralrpc"
	"distributed2048/util"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
)

const (
	ERROR_LOG bool = true
	DEBUG_LOG bool = true
)

var LOGV = util.NewLogger(DEBUG_LOG, "DEBUG", os.Stdout)
var LOGE = util.NewLogger(ERROR_LOG, "ERROR", os.Stderr)

type gameServer struct {
	info centralrpc.Node
}

type centralServer struct {
	nextGameServerID     uint32
	gameServersLock      sync.Mutex
	gameServers          map[uint32]*gameServer
	hostPortToGameServer map[string]*gameServer
	gameServersSlice     []centralrpc.Node
	numGameServers       int
}

func NewCentralServer(port, numGameServers int) (CentralServer, error) {
	cs := &centralServer{
		numGameServers:   numGameServers,
		gameServers:      make(map[uint32]*gameServer),
		gameServersSlice: nil,
	}

	rpc.RegisterName("CentralServer", centralrpc.Wrap(cs))
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	go http.Serve(l, nil)

	return cs, nil
}

func (cs *centralServer) GetGameServerForClient(args *centralrpc.GetGameServerForClientArgs, reply *centralrpc.GetGameServerForClientReply) error {
	return errors.New("Not implemented")
}

func (cs *centralServer) RegisterGameServer(args *centralrpc.RegisterGameServerArgs, reply *centralrpc.RegisterGameServerReply) error {
	cs.gameServersLock.Lock()

	var id uint32
	if gs, exists := cs.hostPortToGameServer[args.HostPort]; !exists {

		// Get a new ID
		id = cs.nextGameServerID
		cs.nextGameServerID++

		// Get host:port
		hostport := args.HostPort

		// Add new server object to map
		gs = &gameServer{centralrpc.Node{id, hostport}}
		cs.gameServers[id] = gs
		cs.hostPortToGameServer[hostport] = gs
	} else {
		id = gs.info.NodeID
	}

	// Check if all the game servers in the ring have registered. If they
	// haven't, then reply with not ready. Otherwise, reply with OK, send back
	// to the unique ID, and the list of all game servers.
	if len(cs.gameServers) < cs.numGameServers {
		reply.Status = centralrpc.NotReady
	} else {
		reply.Status = centralrpc.OK
		reply.GameServerID = id
		// Check if the game servers sliced has been cached. If it hasn't, make it.
		if cs.gameServersSlice == nil {
			cs.gameServersSlice = make([]centralrpc.Node, len(cs.gameServers))
			i := 0
			for _, node := range cs.gameServers {
				cs.gameServersSlice[i] = node.info
				i++
			}
		}
		reply.Servers = cs.gameServersSlice
	}

	cs.gameServersLock.Unlock()

	return nil
}
