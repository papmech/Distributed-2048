package libpaxos

import (
	"distributed2048/rpc/paxosrpc"
)

type PaxosNode struct {
	ID       uint32
	HostPort string
}

type Libpaxos interface {
	ReceiveProposal(args *paxosrpc.ReceiveProposalArgs, reply *paxosrpc.ReceiveProposalReply) error
	ReceiveCommit(args *paxosrpc.ReceiveCommitArgs, reply *paxosrpc.ReceiveCommitReply) error
	ReceiveDecide(args *paxosrpc.ReceiveDecideArgs, reply *paxosrpc.ReceiveDecideReply) error
	Propose(interface{}) error
	DecidedHandler(func(interface{}) error)
}
