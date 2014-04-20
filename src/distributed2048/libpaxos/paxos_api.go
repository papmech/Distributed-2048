package libpaxos

import (
	"distributed2048/gameserver"
	"distributed2048/rpc/paxosrpc"
)

type PaxosNode struct {
	ID       uint32
	HostPort string
}

type Libpaxos interface {
	ReceiveProposal(*paxosrpc.ReceiveProposalArgs, *paxosrpc.ReceiveProposalReply) error
	ReceiveCommit(*paxosrpc.ReceiveCommitArgs, *paxosrpc.ReceiveCommitReply) error
	ReceiveDecide(*paxosrpc.ReceiveDecideArgs, *paxosrpc.ReceiveDecideReply) error
	Propose([]gameserver.Move) error
	DecidedHandler(func(interface{}) error)
}
