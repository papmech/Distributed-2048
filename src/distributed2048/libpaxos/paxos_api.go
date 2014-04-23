package libpaxos

import (
	"distributed2048/rpc/paxosrpc"
)

// Libpaxos defines the methods that a game server can call to propose new
// values AND handle decided values from successful Paxos rounds.
type Libpaxos interface {
	// Propose will queue a new value to be proposed to the rest of the nodes.
	// It will not block.
	Propose(moves []paxosrpc.Move) error
	// DecidedHandler sets the callback function that will be invoked when a
	// Paxos round has completed and a new value has been decided upon.
	DecidedHandler(handler func(slotNumber uint32, moves []paxosrpc.Move))
}
