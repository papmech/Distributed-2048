package libpaxos

import (
	"distributed2048/gameserver"
	"distributed2048/rpc/paxosrpc"
)

type Libpaxos interface {
	// ReceiveProposal is called by a Proposer via RPC when it wishes to
	// propose a new value for Paxos.
	ReceivePrepare(*paxosrpc.ReceivePrepareArgs, *paxosrpc.ReceivePrepareReply) error
	// ReceiveAccept is called by a Proposer via RPC when it wishes to ask all
	// other nodes if they accept the proposed value.
	ReceiveAccept(*paxosrpc.ReceiveAcceptArgs, *paxosrpc.ReceiveAcceptReply) error
	// ReceiveDecide is called by a Proposer via RPC when one round of Paxos
	// has completed, and everyone has agreed on a value.
	ReceiveDecide(*paxosrpc.ReceiveDecideArgs, *paxosrpc.ReceiveDecideReply) error
	// Propose will propose a new value among the other nodes. It will block
	// until the proposal has been successfully accepted.
	Propose([]gameserver.Move) error
	// DecidedHandler(func([]gameserver.Move) error)
}
