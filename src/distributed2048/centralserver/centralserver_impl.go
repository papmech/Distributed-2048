package centralserver

import (
	"distributed2048/rpc/centralrpc"
	"distributed2048/util"
	"errors"
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
	nextGameServerID uint32
	gameServersLock  sync.Mutex
	gameServers      map[uint32]*gameServer
	numGameServers   int
}

func NewCentralServer(port, numGameServers int) (CentralServer, error) {
	something := &centralServer{
		numGameServers: numGameServers,
		gameServers:    make(map[uint32]*gameServer),
	}

	return something, nil
}

func (cs *centralServer) GetGameServerForClient(args *centralrpc.GetGameServerForClientArgs, reply *centralrpc.GetGameServerForClientReply) error {
	return errors.New("Not implemented")
}

func (cs *centralServer) RegisterGameServer(args *centralrpc.RegisterGameServerArgs, reply *centralrpc.RegisterGameServerReply) error {
	return errors.New("Not implemented")
}
