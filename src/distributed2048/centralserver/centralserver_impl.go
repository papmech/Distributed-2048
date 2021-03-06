package centralserver

import (
	"distributed2048/rpc/centralrpc"
	"distributed2048/rpc/paxosrpc"
	"distributed2048/util"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
)

const (
	ERROR_LOG bool = true
	DEBUG_LOG bool = false
)

var LOGV = util.NewLogger(DEBUG_LOG, "DEBUG", os.Stdout)
var LOGE = util.NewLogger(ERROR_LOG, "ERROR", os.Stderr)

type gameServer struct {
	info        paxosrpc.Node
	clientCount int
}

type centralServer struct {
	nextGameServerID     uint32
	gameServersLock      sync.Mutex
	gameServers          map[uint32]*gameServer
	hostPortToGameServer map[string]*gameServer
	gameServersSlice     []paxosrpc.Node
	numGameServers       int
}

func NewCentralServer(port, numGameServers int) (CentralServer, error) {
	LOGV.Println("New Central Server is starting up")
	if numGameServers < 1 {
		return nil, errors.New("numGameServers must be at least 1")
	}

	cs := &centralServer{
		numGameServers:       numGameServers,
		gameServers:          make(map[uint32]*gameServer),
		hostPortToGameServer: make(map[string]*gameServer),
		gameServersSlice:     nil,
	}

	// Serve up information for the game client

	http.HandleFunc("/", cs.gameClientViewHandler)
	go http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

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
	cs.gameServersLock.Lock()
	if len(cs.gameServers) < cs.numGameServers {
		// Not all game servers have connected to the ring, so reply with NotReady
		reply.Status = centralrpc.NotReady
	} else {
		id := cs.getGameServerIDMinClients()
		cs.gameServers[id].clientCount++
		reply.Status = centralrpc.OK
		reply.HostPort = cs.gameServers[id].info.HostPort
	}
	cs.gameServersLock.Unlock()

	return nil
}

func (cs *centralServer) RegisterGameServer(args *centralrpc.RegisterGameServerArgs, reply *centralrpc.RegisterGameServerReply) error {
	cs.gameServersLock.Lock()

	var id uint32
	if gs, exists := cs.hostPortToGameServer[args.HostPort]; !exists {

		// Check if we hit the limit
		if len(cs.gameServers) >= cs.numGameServers {
			LOGE.Println("Received request to register game server when ring is FULL")
			reply.Status = centralrpc.Full
			cs.gameServersLock.Unlock()
			return nil
		}

		// Get a new ID
		id = cs.nextGameServerID
		cs.nextGameServerID++

		// Get host:port
		hostport := args.HostPort

		// Add new server object to map
		gs = &gameServer{paxosrpc.Node{id, hostport}, 0}
		cs.gameServers[id] = gs
		cs.hostPortToGameServer[hostport] = gs
	} else {
		id = gs.info.ID
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
			cs.gameServersSlice = make([]paxosrpc.Node, len(cs.gameServers))
			i := 0
			for _, node := range cs.gameServers {
				cs.gameServersSlice[i] = node.info
				i++
			}
		}
		reply.Servers = cs.gameServersSlice
	}

	LOGV.Printf("Received registration request from %d, reply was %d\n", id, reply.Status)

	cs.gameServersLock.Unlock()

	return nil
}

type HttpReply struct {
	Status   string
	Hostport string
}

func (cs *centralServer) gameClientViewHandler(w http.ResponseWriter, r *http.Request) {
	LOGV.Println("a new request was made with URI " + r.RequestURI)
	reply := HttpReply{}
	cs.gameServersLock.Lock()
	if len(cs.gameServers) < cs.numGameServers {
		// Not all game servers have connected to the ring, so reply with NotReady
		LOGV.Println("Not all game servers have connected - replying not ready...")
		reply.Status = "NotReady"
		reply.Hostport = ""
	} else {
		LOGV.Println("Games servers have connected - replying with OK")
		id := cs.getGameServerIDMinClients()
		cs.gameServers[id].clientCount++
		reply.Status = "OK"
		reply.Hostport = cs.gameServers[id].info.HostPort
	}
	cs.gameServersLock.Unlock()
	buf, err := json.Marshal(reply)
	if err == nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Connection", "Keep-Alive")
		_, err = w.Write(buf)
	} else {
		LOGE.Printf("Error with marshalling reply: " + err.Error())
	}
}

func (cs *centralServer) getGameServerIDMinClients() uint32 {
	// Must be called with the LOCK acquired
	min := math.MaxInt32
	var resultID uint32
	for _, gs := range cs.gameServers {
		if gs.clientCount < min {
			min = gs.clientCount
			resultID = gs.info.ID
		}
	}
	return resultID
}
