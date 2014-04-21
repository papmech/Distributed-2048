package libpaxos

import (
	"distributed2048/rpc/paxosrpc"
)

type Libpaxos interface {
	// ReceiveProposal is called by a Proposer via RPC when it wishes to
	// propose a new value for Paxos.
	ReceivePrepare(args *paxosrpc.ReceivePrepareArgs, reply *paxosrpc.ReceivePrepareReply) error
	// ReceiveAccept is called by a Proposer via RPC when it wishes to ask all
	// other nodes if they accept the proposed value.
	ReceiveAccept(args *paxosrpc.ReceiveAcceptArgs, reply *paxosrpc.ReceiveAcceptReply) error
	// ReceiveDecide is called by a Proposer via RPC when one round of Paxos
	// has completed, and everyone has agreed on a value.
	ReceiveDecide(args *paxosrpc.ReceiveDecideArgs, reply *paxosrpc.ReceiveDecideReply) error
	// Propose will queue a new value to be proposed to the rest of the nodes.
	// It will not block.
	Propose(moves []paxosrpc.Move) error
	// DecidedHandler sets the callback function that will be invoked when a
	// Paxos round has completed and a new value has been decided upon.
	DecidedHandler(func([]paxosrpc.Move))
}
