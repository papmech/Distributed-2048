// This is the api for a 2048 game server.

package gameserver

import (
	"distributed2048/rpc/paxosrpc"
)

type GameServer interface {
	DoVote()
	AddVote()
	SetVoteResult()
	TestAddVote(moves []paxosrpc.Move)
	ListenForClients()
}
