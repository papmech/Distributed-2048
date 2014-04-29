// This is the api for a 2048 game server.

package gameserver

import (
	"distributed2048/libpaxos"
)

type GameServer interface {
	ListenForClients()
	GetLibpaxos() libpaxos.Libpaxos
}
