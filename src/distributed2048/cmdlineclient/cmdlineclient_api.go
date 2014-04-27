// A simple 2048 client running on a command line that facilitates automated testing
package cmdlineclient

import(
	"distributed2048/lib2048"
	"distributed2048/rpc/paxosrpc"
)

type Cclient interface {
	InputMove(move paxosrpc.Direction)
	GetGameState() (lib2048.Grid, int, bool, bool)
	Close()
}
