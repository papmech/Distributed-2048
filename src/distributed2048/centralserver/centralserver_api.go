package centralserver

import (
	"distributed2048/rpc/centralrpc"
)

type CentralServer interface {
	GetGameServerForClient(args *centralrpc.GetGameServerForClientArgs, reply *centralrpc.GetGameServerForClientReply)
}
