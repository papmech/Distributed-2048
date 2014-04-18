package centralserver

import (
	"distributed2048/rpc/centralrpc"
)

type CentralServer interface {
	GetGameServerForClient(args *centralrpc.GetGameServerForClientArgs, reply *centralrpc.GetGameServerForClientReply) error
	RegisterGameServer(args *centralrpc.RegisterGameServerArgs, reply *centralrpc.RegisterGameServerReply) error
}
