package centralrpc

type Status int

const (
	OK Status = iota + 1 // RPC was a success
)

type GetGameServerForClientArgs struct {
	// nothing here
}

type GetGameServerForClientReply struct {
	Status   Status
	HostPort string // Host:Port of the assigned game server
}
