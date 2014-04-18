package centralrpc

type Status int

const (
	OK Status = iota + 1 // RPC was a success
	NotReady
)

type Node struct {
	HostPort string
	NodeID   uint32
}

type GetGameServerForClientArgs struct {
	// nothing here
}

type GetGameServerForClientReply struct {
	Status   Status
	HostPort string // Host:Port of the assigned game server
}

type RegisterGameServerArgs struct {
	HostPort string // Host:Port of the registering game server
}

type RegisterGameServerReply struct {
	Status       Status
	GameServerID int // Unique ID
}
