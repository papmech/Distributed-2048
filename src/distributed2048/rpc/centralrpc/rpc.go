package centralrpc

type RemoteCentralServer interface {
	GetGameServerForClient(*GetGameServerForClientArgs, *GetGameServerForClientReply) error
	RegisterGameServer(*RegisterGameServerArgs, *RegisterGameServerReply) error
}

type CentralServer struct {
	RemoteCentralServer
}

// Wrap wraps s in a type-safe wrapper struct to ensure that only the desired
// StorageServer methods are exported to receive RPCs.
func Wrap(cs RemoteCentralServer) RemoteCentralServer {
	return &CentralServer{cs}
}
