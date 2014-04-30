package libpaxos

import (
	"distributed2048/rpc/paxosrpc"
)

type PaxosAction int

const (
	Prepare PaxosAction = iota + 1
	Accept
	Decide
)

// Libpaxos defines the methods that a game server can call to propose new
// values AND handle decided values from successful Paxos rounds.
type Libpaxos interface {
	// Propose will queue a new value to be proposed to the rest of the nodes.
	// It will not block.
	Propose(*paxosrpc.ProposalValue) error
	// DecidedHandler sets the callback function that will be invoked when a
	// Paxos round has completed and a new value has been decided upon.
	DecidedHandler(handler func(proposal *paxosrpc.ProposalValue))
	// SetInterruptFunc sets the function that will be called at the beginning
	// of every Paxos related receiving step (i.e. ReceivePrepare,
	// ReceiveAccept, ReceiveDecide). This is useful for inserting debugging
	// or testing code, to lag / interrupt the server for example.
	SetInterruptFunc(f func(id uint32, action PaxosAction, slotNumber uint32))
}
