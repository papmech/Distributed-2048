// This is the api for a 2048 game server.

package gameserver

import (

)

type GameServer interface {
	DoVote() ()
	AddVote() ()
	SetVoteResult() ()
	NewServer(pattern string) *GameServer
	ListenForClients()

}

