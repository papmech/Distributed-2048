package centralserver

import (
	"distributed2048/rpc/centralrpc"
)

type CentralServer interface {
	// GetGameServerForClient returns the server that a game client can
	// connect to. It replies with NotReady if not all the game servers have
	// joined the ring. Once all game servers have connected, it replies with
	// OK and information about a single game server that the client should
	// connect to. This server is chosen so that the load across the game
	// servers is balanced. This function is called when the game client
	// initially wishes to connect, and also when it has lost its existing
	// connection and wishes to reconnect.
	GetGameServerForClient(args *centralrpc.GetGameServerForClientArgs, reply *centralrpc.GetGameServerForClientReply) error

	// RegisterGameServer adds a game server to the ring. It replies with
	// NotReady if not all the game servers have joined the ring. Once all
	// game servers have joined, it replies with OK and a list of all the game
	// servers in the ring.
	RegisterGameServer(args *centralrpc.RegisterGameServerArgs, reply *centralrpc.RegisterGameServerReply) error
}
