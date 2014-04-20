package centralrpc

import (
	"distributed2048/rpc/paxosrpc"
)

type Status int

const (
	OK Status = iota + 1 // RPC was a success
	NotReady
	Full
)

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
	GameServerID uint32 // Unique ID
	Servers      []paxosrpc.Node
}
